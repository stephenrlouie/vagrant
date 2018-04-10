#!/bin/bash

# Get root access.
sudo su

# Disable swap.
swapoff -a

# Get the latest.
yum -y update

# Install Docker.
yum install -y docker-2:1.13.1-53.git774336d.el7.centos.x86_64
systemctl enable docker && systemctl start docker

# Install kubeadm.
cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF
setenforce 0
yum install -y kubelet-1.10.0-0.x86_64 kubeadm-1.10.0-0.x86_64 kubectl-1.10.0-0.x86_64
systemctl enable kubelet && systemctl start kubelet
cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sysctl --system

# Initialize the master node.
echo "Running kubeadm init..."
JOIN_CMD=$(kubeadm init --kubernetes-version 1.10.0 | grep "kubeadm join")
echo "$JOIN_CMD"

# Allow kubectl to work for non-root users.
mkdir -p /home/vagrant/.kube
cp /etc/kubernetes/admin.conf /home/vagrant/.kube/config
chown vagrant:vagrant /home/vagrant/.kube/config

# Export the admin config.
export KUBECONFIG=/home/vagrant/.kube/config

# Configure Weave Net CNI pod networking.
sysctl net.bridge.bridge-nf-call-iptables=1
export kubever=$(kubectl version | base64 | tr -d '\n')
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$kubever"

# Allow scheduling on the master node.
kubectl taint nodes --all node-role.kubernetes.io/master-

# Join nodes.
eval "$JOIN_CMD --ignore-preflight-errors=all"

# Make sure the machine has the KUBECONFIG exported when it starts up.
echo "export KUBECONFIG=/home/vagrant/.kube/config" > /etc/profile.d/kube-provision.sh
chmod 0755 /etc/profile.d/kube-provision.sh
