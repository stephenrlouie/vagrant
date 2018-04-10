#!/bin/bash

sudo su
swapoff -a
setenforce 0
sysctl --system
kubeadm reset
echo "Running kubeadm init..."
JOIN_CMD=$(kubeadm init --apiserver-advertise-address ${MYIP} --pod-network-cidr 10.1.0.0/16 | grep "kubeadm join")
echo "$JOIN_CMD"
cp /etc/kubernetes/admin.conf /home/vagrant/.kube/config
chown vagrant:vagrant /home/vagrant/.kube/config
sysctl net.bridge.bridge-nf-call-iptables=1
export kubever=$(kubectl version | base64 | tr -d '\n')
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$kubever"
kubectl taint nodes --all node-role.kubernetes.io/master-
eval "$JOIN_CMD --ignore-preflight-errors=all"
