#!/bin/bash
sed -i '/nameserver 10.0.2.3/i nameserver 172.16.7.104' /etc/resolv.conf
