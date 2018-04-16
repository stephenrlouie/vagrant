#!/bin/sh

sudo su

python /home/vagrant/inject-kubeconfig.py $1


curl -X POST \
  http://172.16.7.101:30900/v0/clusters \
  -H 'Content-Type: application/json' \
  -d @/home/vagrant/my-cluster.json
