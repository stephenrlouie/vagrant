#!/bin/bash

# enable root access
sudo su

# create a persistent volume
kubectl create -f /home/vagrant/pv1.yaml
kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default

# use helm client CLI to deploy helm tiller onto this cluster
helm init
