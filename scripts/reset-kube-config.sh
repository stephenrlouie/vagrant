#!/bin/bash

# Get root access.
sudo su

# Disable swap.
swapoff -a

# Re-apply kubeadm settings.
setenforce 0
sysctl --system

# Reset kubeadm.
kubeadm reset

# Re-init kubeadm.
echo "Running kubeadm init..."
JOIN_CMD=$(kubeadm init --feature-gates CoreDNS=true --kubernetes-version 1.10.0 --apiserver-advertise-address ${MYIP} --pod-network-cidr 10.1.0.0/16 | grep "kubeadm join")
echo "$JOIN_CMD"

# Allow kubectl to work for non-root users.
cp /etc/kubernetes/admin.conf /home/vagrant/.kube/config
chown vagrant:vagrant /home/vagrant/.kube/config

# Configure Weave Net CNI pod networking.
sysctl net.bridge.bridge-nf-call-iptables=1
export kubever=$(kubectl version | base64 | tr -d '\n')
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$kubever"

# Allow scheduling on master.
kubectl taint nodes --all node-role.kubernetes.io/master-

# Kubeadm join without preflight checks.
eval "$JOIN_CMD --ignore-preflight-errors=all"
