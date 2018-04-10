#!/bin/bash

sudo su

mkdir -p /etc/kubernetes/edge
kubectl create -f /home/vagrant/pv2.yaml

PACKAGE=client
VERSION=20180402
curl -O http://storage.googleapis.com/crreleases/nightly/$VERSION/clusterregistry-$PACKAGE.tar.gz
tar xzf clusterregistry-$PACKAGE.tar.gz

cp ./crinit /usr/bin

crinit aggregated init optikon-cr  --host-cluster-context=$(kubectl config current-context)
