#!/usr/bin/env bash

set -e
set -x

for i in `seq 0 7`; do
    iptables -A INPUT -p tcp --dport 8080 -s 192.168.1.$i -j ACCEPT
done

iptables -A INPUT -p tcp --dport 8080 -s 127.0.0.0/8 -j ACCEPT

iptables -P INPUT DROP
