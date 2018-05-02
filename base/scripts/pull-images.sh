#!/bin/bash
docker pull gcr.io/crreleases/clusterregistry:latest
docker pull k8s.gcr.io/etcd:3.0.17
docker pull nginx:1.7.9
docker pull prom/alertmanager:v0.14.0
docker pull jimmidyson/configmap-reload:v0.1
docker pull busybox:latest
docker pull k8s.gcr.io/kube-state-metrics:v1.2.0
docker pull prom/node-exporter:v0.15.2
docker pull prom/prometheus:v2.2.1
docker pull prom/pushgateway:v0.4.0
docker pull dockerhub.cisco.com/intelligent-edge-dev-docker-local/optikon-ui:0.1
docker pull dockerhub.cisco.com/intelligent-edge-dev-docker-local/optikon-api:0.1