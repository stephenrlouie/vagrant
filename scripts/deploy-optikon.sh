# !/bin/bash

sudo su


# Deploy API --> port 30900
kubectl create configmap central-kubeconfig --from-file /etc/kubernetes/admin.conf
kubectl create -f /home/vagrant/optikon-api.yaml


# Deploy UI --> port 30800
kubectl create -f /home/vagrant/optikon-ui.yaml
