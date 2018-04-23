/*
 * Copyright 2018 The CoreDNS Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * NOTE: This software contains code derived from the the Apache-licensed CoreDNS
 * `forward` plugin (https://github.com/coredns/coredns/tree/master/plugin/forward),
 * including various modifications by Cisco Systems, Inc.
 */

package edge

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"wwwin-github.cisco.com/edge/optikon-dns/plugin/central"

	"github.com/miekg/dns"
	ot "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
)

// OptikonEdge represents a plugin instance that can proxy requests to another (DNS) server. It has a list
// of proxies each representing one upstream proxy.
type OptikonEdge struct {
	proxies    []*Proxy
	p          Policy
	hcInterval time.Duration

	from    string
	ignored []string

	tlsConfig     *tls.Config
	tlsServerName string
	maxfails      uint32
	expire        time.Duration

	forceTCP bool // also here for testing

	Next plugin.Handler

	lon      float64
	lat      float64
	services []string
}

// New returns a new OptikonEdge.
func New() *OptikonEdge {
	oe := &OptikonEdge{maxfails: 2, tlsConfig: new(tls.Config), expire: defaultExpire, p: new(random), from: ".", hcInterval: hcDuration}
	return oe
}

// SetProxy appends p to the proxy list and starts healthchecking.
func (oe *OptikonEdge) SetProxy(p *Proxy) {
	oe.proxies = append(oe.proxies, p)
	p.start(oe.hcInterval)
}

// Len returns the number of configured proxies.
func (oe *OptikonEdge) Len() int { return len(oe.proxies) }

// Name implements plugin.Handler.
func (oe *OptikonEdge) Name() string { return "optikon-edge" }

// ServeDNS implements plugin.Handler.
func (oe *OptikonEdge) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	state := request.Request{W: w, Req: r}
	if !oe.match(state) {
		return plugin.NextOrFailure(oe.Name(), oe.Next, ctx, w, r)
	}

	fails := 0
	var span, child ot.Span
	var upstreamErr error
	span = ot.SpanFromContext(ctx)

	for _, proxy := range oe.list() {
		if proxy.Down(oe.maxfails) {
			fails++
			if fails < len(oe.proxies) {
				continue
			}
			// All upstream proxies are dead, assume healtcheck is completely broken and randomly
			// select an upstream to connect to.
			r := new(random)
			proxy = r.List(oe.proxies)[0]

			HealthcheckBrokenCount.Add(1)
		}

		if span != nil {
			child = span.Tracer().StartSpan("connect", ot.ChildOf(span.Context()))
			ctx = ot.ContextWithSpan(ctx, child)
		}

		var (
			ret *dns.Msg
			err error
		)
		stop := false
		for {
			ret, err = proxy.connect(ctx, state, oe.forceTCP, true)
			if err != nil && err == io.EOF && !stop { // Remote side closed conn, can only happen with TCP.
				stop = true
				continue
			}
			break
		}

		if child != nil {
			child.Finish()
		}

		ret, err = truncated(ret, err)
		upstreamErr = err

		if err != nil {
			// Kick off health check to see if *our* upstream is broken.
			if oe.maxfails != 0 {
				proxy.Healthcheck()
			}

			if fails < len(oe.proxies) {
				continue
			}
			break
		}

		// Check if the reply is correct; if not return FormErr.
		if !state.Match(ret) {
			formerr := state.ErrorMessage(dns.RcodeFormatError)
			w.WriteMsg(formerr)
			return 0, nil
		}

		ret.Compress = true
		// When using force_tcp the upstream can send a message that is too big for
		// the udp buffer, hence we need to truncate the message to at least make it
		// fit the udp buffer.
		ret, _ = state.Scrub(ret)

		// Assert an additional entry for the table exists.
		if len(ret.Extra) == 0 {
			if len(ret.Answer) == 0 {
				fmt.Println("ERROR: No Additional entries returned!")
				return dns.RcodeServerFailure, errTableParseFailure
			}
			w.WriteMsg(ret)
			return 0, nil
		}

		// Extract the edge sites from the response.
		edgeSiteRR := ret.Extra[0]
		edgeSiteSubmatches := edgeSiteRegex.FindStringSubmatch(edgeSiteRR.String())
		if len(edgeSiteSubmatches) < 2 {
			fmt.Println("EDGESITERR:", edgeSiteRR)
			fmt.Println("EDGESITESUBMATCHES:", edgeSiteSubmatches)
			return dns.RcodeServerFailure, errTableParseFailure
		}
		edgeSiteStr, err := strconv.Unquote(fmt.Sprintf("\"%s\"", edgeSiteSubmatches[1]))
		if err != nil {
			fmt.Println("EDGESITESTR:", edgeSiteStr)
			return dns.RcodeServerFailure, errTableParseFailure
		}
		var edgeSites []central.EdgeSite
		if err := json.Unmarshal([]byte(edgeSiteStr), &edgeSites); err != nil {
			fmt.Println("EDGESITESTR:", edgeSiteStr)
			return dns.RcodeServerFailure, errTableParseFailure
		}

		// Remove the Table entry from the return message.
		ret.Extra = ret.Extra[1:]

		// If the list is empty, call the next plugin (proxy).
		if len(edgeSites) == 0 {
			return plugin.NextOrFailure(oe.Name(), oe.Next, ctx, w, r)
		}

		// Compute the distance to the first edge site.
		closest := edgeSites[0].IP
		minDist := Distance(oe.lat, oe.lon, edgeSites[0].Lat, edgeSites[0].Lon)
		for _, edgeSite := range edgeSites {
			dist := Distance(oe.lat, oe.lon, edgeSite.Lat, edgeSite.Lon)
			if dist < minDist {
				minDist = dist
				closest = edgeSite.IP
			}
		}

		// Write the closest cluster IP as a DNS record.
		var rr dns.RR
		switch state.Family() {
		case 1:
			rr = new(dns.A)
			rr.(*dns.A).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass()}
			rr.(*dns.A).A = net.ParseIP(closest).To4()
		case 2:
			rr = new(dns.AAAA)
			rr.(*dns.AAAA).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeAAAA, Class: state.QClass()}
			rr.(*dns.AAAA).AAAA = net.ParseIP(closest)
		}
		ret.Answer = []dns.RR{rr}

		// Write the response message.
		w.WriteMsg(ret)

		return 0, nil
	}

	if upstreamErr != nil {
		return dns.RcodeServerFailure, upstreamErr
	}

	return dns.RcodeServerFailure, errNoHealthy
}

func (oe *OptikonEdge) match(state request.Request) bool {
	from := oe.from

	if !plugin.Name(from).Matches(state.Name()) || !oe.isAllowedDomain(state.Name()) {
		return false
	}

	return true
}

func (oe *OptikonEdge) isAllowedDomain(name string) bool {
	if dns.Name(name) == dns.Name(oe.from) {
		return true
	}

	for _, ignore := range oe.ignored {
		if plugin.Name(ignore).Matches(name) {
			return false
		}
	}
	return true
}

// List returns a set of proxies to be used for this client depending on the policy in oe.
func (oe *OptikonEdge) list() []*Proxy { return oe.p.List(oe.proxies) }

var (
	errInvalidDomain         = errors.New("invalid domain for forward")
	errNoHealthy             = errors.New("no healthy proxies")
	errNoOptikonEdge         = errors.New("no optikon-edge defined")
	errTableParseFailure     = errors.New("unable to parse Table returned from central")
	errFindingClosestCluster = errors.New("unable to compute closest edge cluster")
)

// policy tells forward what policy for selecting upstream it uses.
type policy int

const (
	randomPolicy policy = iota
	roundRobinPolicy
)

var (
	edgeSiteRegex = regexp.MustCompile(`^.*\t0\tIN\tTXT\t\"(\[.*\])\"$`)
)
