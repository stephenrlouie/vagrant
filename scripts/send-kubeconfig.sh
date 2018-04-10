#!/bin/bash

chmod 700 /home/vagrant/id_rsa
scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i /home/vagrant/id_rsa  /etc/kubernetes/admin.conf vagrant@172.16.7.101:/home/vagrant/.edge-kubeconfig/edge-$1.conf
