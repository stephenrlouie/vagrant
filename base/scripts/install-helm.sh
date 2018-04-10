#!/bin/bash
# download helm (client CLI) binary
wget https://storage.googleapis.com/kubernetes-helm/helm-v2.8.2-linux-amd64.tar.gz
tar -zxvf helm-v2.8.2-linux-amd64.tar.gz
cp linux-amd64/helm /usr/bin/
