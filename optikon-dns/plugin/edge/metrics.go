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
 */

package edge

import (
	"sync"

	"github.com/coredns/coredns/plugin"

	"github.com/prometheus/client_golang/prometheus"
)

// Variables declared for monitoring.
var (
	RequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "optikon-edge",
		Name:      "request_count_total",
		Help:      "Counter of requests made per upstream.",
	}, []string{"to"})
	RcodeCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "optikon-edge",
		Name:      "response_rcode_count_total",
		Help:      "Counter of requests made per upstream.",
	}, []string{"rcode", "to"})
	RequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: plugin.Namespace,
		Subsystem: "optikon-edge",
		Name:      "request_duration_seconds",
		Buckets:   plugin.TimeBuckets,
		Help:      "Histogram of the time each request took.",
	}, []string{"to"})
	HealthcheckFailureCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "optikon-edge",
		Name:      "healthcheck_failure_count_total",
		Help:      "Counter of the number of failed healtchecks.",
	}, []string{"to"})
	HealthcheckBrokenCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "optikon-edge",
		Name:      "healthcheck_broken_count_total",
		Help:      "Counter of the number of complete failures of the healtchecks.",
	})
	SocketGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: plugin.Namespace,
		Subsystem: "optikon-edge",
		Name:      "socket_count_total",
		Help:      "Gauge of open sockets per upstream.",
	}, []string{"to"})
)

var once sync.Once
