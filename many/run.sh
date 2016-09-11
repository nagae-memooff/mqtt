#!/bin/sh
ulimit -n 1048576

sudo ifconfig eth1 192.168.100.110 up
sudo ifconfig eth2 192.168.100.111 up
sudo ifconfig eth3 192.168.100.112 up
sudo ifconfig eth4 192.168.100.113 up
sudo ifconfig eth0:1 192.168.100.114 up
sudo ifconfig eth1:1 192.168.100.115 up
sudo ifconfig eth2:1 192.168.100.116 up
sudo ifconfig eth3:1 192.168.100.117 up
sudo ifconfig eth4:1 192.168.100.118 up


./many -host 192.168.100.228:1884 -laddr=192.168.100.227 -wait=100 -close_wait=5 -pace=15000 -start_port=1024 -ssl=false -start_wait=200 -list_file=channel.list.0 -t=true  -ping_only=true &
./many -host 192.168.100.228:1884 -laddr=192.168.100.110 -wait=100 -close_wait=5 -pace=15000 -start_port=1024 -ssl=false -start_wait=200 -list_file=channel.list.1 -t=false -ping_only=true &
./many -host 192.168.100.228:1884 -laddr=192.168.100.111 -wait=100 -close_wait=5 -pace=15000 -start_port=1024 -ssl=false -start_wait=200 -list_file=channel.list.2 -t=false -ping_only=true &
./many -host 192.168.100.228:1884 -laddr=192.168.100.112 -wait=100 -close_wait=5 -pace=15000 -start_port=1024 -ssl=false -start_wait=200 -list_file=channel.list.3 -t=false -ping_only=true &
./many -host 192.168.100.228:1884 -laddr=192.168.100.113 -wait=100 -close_wait=5 -pace=15000 -start_port=1024 -ssl=false -start_wait=200 -list_file=channel.list.4 -t=false -ping_only=true &
./many -host 192.168.100.228:1884 -laddr=192.168.100.114 -wait=100 -close_wait=5 -pace=15000 -start_port=1024 -ssl=false -start_wait=200 -list_file=channel.list.5 -t=false -ping_only=true &
./many -host 192.168.100.228:1884 -laddr=192.168.100.115 -wait=100 -close_wait=5 -pace=15000 -start_port=1024 -ssl=false -start_wait=200 -list_file=channel.list.6 -t=false -ping_only=true &
./many -host 192.168.100.228:1884 -laddr=192.168.100.116 -wait=100 -close_wait=5 -pace=15000 -start_port=1024 -ssl=false -start_wait=200 -list_file=channel.list.7 -t=false -ping_only=true &
./many -host 192.168.100.228:1884 -laddr=192.168.100.117 -wait=100 -close_wait=5 -pace=15000 -start_port=1024 -ssl=false -start_wait=200 -list_file=channel.list.8 -t=false -ping_only=true &
./many -host 192.168.100.228:1884 -laddr=192.168.100.118 -wait=100 -close_wait=5 -pace=15000 -start_port=1024 -ssl=false -start_wait=200 -list_file=channel.list.9 -t=false -ping_only=true &
