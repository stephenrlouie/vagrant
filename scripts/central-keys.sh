#!/bin/bash

key=`cat /home/vagrant/id_rsa.pub`

for i in $(seq 1 $NUM_EDGE); do
  temp=$key$i;
  echo $temp >> ~/.ssh/authorized_keys;
done

mkdir -p /home/vagrant/.edge-kubeconfig
