#!/bin/bash

sudo su

# cluster registry needs a persistent volume
kubectl create -f /home/vagrant/pv.yaml

# download cluster registry
PACKAGE=client
VERSION=20180402
curl -O http://storage.googleapis.com/crreleases/nightly/$VERSION/clusterregistry-$PACKAGE.tar.gz
tar xzf clusterregistry-$PACKAGE.tar.gz

# crinit is the cluster registry CLI. add to path...
cp ./crinit /usr/bin

# use the crinit CLI helper to deploy the c-registry aggregated API server, on this cluster
crinit aggregated init optikon-cr  --host-cluster-context=$(kubectl config current-context)
