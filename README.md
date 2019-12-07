# Compile

```shell script
bash ./script/build.sh
```

# Run

```shell script
./vpn_linux_amd64 ${PEER_IP} ${PEER_HOST} ${VPN_NETWORK}
```

# Example

Using virtualBox for example(`CentOS-7-x86_64-Minimal-1908.iso`)

1. Set `NAT network`(not `NAT`)
* In this example program, we use TCP channel to transmit IP packets, so we must ensure that the two virtual machine networks are connected

virtual machine 1(hostname is `vpn1`)
* ip:`10.0.2.6`

virtual machine 2(hostname is `vpn2`)
* ip:`10.0.2.7`

Run the following command in `vpn1`

1. create tun device with local ip `192.169.66.1`
1. route packets with destination IP address' 192.169.66.0/24 'to tun device
1. create tcp channel with peer side(vpn2 10.0.2.7)

```shell script
[root@vpn-1 /]$ ./vpn_linux_amd64 10.0.2.7 9999 192.169.66.1/24

#-------------------------output-------------------------
2019/12/07 09:03:13 tunIp='192.169.66.1'
2019/12/07 09:03:13 Tun Interface Name: tun0
2019/12/07 09:03:13 exec command 'ip address add 192.169.66.1 dev tun0'
2019/12/07 09:03:13 exec command 'ip link set dev tun0 up'
2019/12/07 09:03:13 exec command 'ip route add table main 192.169.66.0/24 dev tun0'
2019/12/07 09:03:13 non ipv4
2019/12/07 09:03:13 listener on '0.0.0.0:9999'
2019/12/07 09:03:13 try to connect peer
2019/12/07 09:03:13 try to reconnect 1s later, addr=10.0.2.7:9999, err=dial tcp 10.0.2.7:9999: connect: connection refused
2019/12/07 09:03:14 try to reconnect 1s later, addr=10.0.2.7:9999, err=dial tcp 10.0.2.7:9999: connect: connection refused
2019/12/07 09:03:14 accept peer success
2019/12/07 09:03:15 connect peer success
#-------------------------output-------------------------
```

Run the following command in `vpn2`

1. create tun device with local ip `192.169.66.2`
1. route packets with destination IP address' 192.169.66.0/24 'to tun device
1. create tcp channel with peer side(vpn2 10.0.2.7)

```shell script
[root@vpn-2 /]$ ./vpn_linux_amd64 10.0.2.6 9999 192.169.66.2/24

#-------------------------output-------------------------
2019/12/07 09:03:14 tunIp='192.169.66.2'
2019/12/07 09:03:14 Tun Interface Name: tun0
2019/12/07 09:03:14 exec command 'ip address add 192.169.66.2 dev tun0'
2019/12/07 09:03:14 exec command 'ip link set dev tun0 up'
2019/12/07 09:03:14 exec command 'ip route add table main 192.169.66.0/24 dev tun0'
2019/12/07 09:03:14 non ipv4
2019/12/07 09:03:14 listener on '0.0.0.0:9999'
2019/12/07 09:03:14 try to connect peer
2019/12/07 09:03:14 connect peer success
2019/12/07 09:03:15 accept peer success
#-------------------------output-------------------------
```

We exec `ping 192.169.66.2 -c 1` on `vpn1(10.0.2.6)`

```shell script
[root@vpn-1 ~]$ ping 192.169.66.2 -c 1

#-------------------------output-------------------------
PING 192.169.66.2 (192.169.66.2) 56(84) bytes of data.
64 bytes from 192.169.66.2: icmp_seq=1 ttl=64 time=3.00 ms

--- 192.169.66.2 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 3.000/3.000/3.000/0.000 ms
#-------------------------output-------------------------
```

Here is the logs from machine `vpn1`

```
2019/12/07 09:04:11 receive from tun, send through tunnel IPFrame {
	version=4
	headerLen=5
	tos=0
	totalLen=84
	identification=17885
	flag=2
	offset=0
	ttl=64
	protocol=1
	headerCheckSum=61301
	source=192.169.66.1
	target=192.169.66.2
	payloadLen=64
}

2019/12/07 09:04:11 receive from tunnel, send through raw socketIPFrame {
	version=4
	headerLen=5
	tos=0
	totalLen=84
	identification=46327
	flag=0
	offset=0
	ttl=64
	protocol=1
	headerCheckSum=49243
	source=192.169.66.2
	target=192.169.66.1
	payloadLen=64
}
```

Here is the logs from machine `vpn2`

```
2019/12/07 09:04:12 receive from tunnel, send through raw socketIPFrame {
	version=4
	headerLen=5
	tos=0
	totalLen=84
	identification=17885
	flag=2
	offset=0
	ttl=64
	protocol=1
	headerCheckSum=61301
	source=192.169.66.1
	target=192.169.66.2
	payloadLen=64
}

2019/12/07 09:04:12 receive from tun, send through tunnel IPFrame {
	version=4
	headerLen=5
	tos=0
	totalLen=84
	identification=46327
	flag=0
	offset=0
	ttl=64
	protocol=1
	headerCheckSum=49243
	source=192.169.66.2
	target=192.169.66.1
	payloadLen=64
}
```