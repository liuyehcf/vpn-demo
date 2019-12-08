package main

import (
	"fmt"
	"github.com/liuyehcf/common-gtools/assert"
	buf "github.com/liuyehcf/common-gtools/buffer"
	"github.com/liuyehcf/vpn-demo/tunnel"
	"github.com/songgao/water"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

var (
	// tcp tunnel's ip and port
	peerIp   net.IP
	peerPort int

	// ip in current side
	tunIp net.IP

	// virtual network
	tunNet *net.IPNet

	// tun interface
	tunIf *water.Interface

	tcpPipe = make(chan []byte)

	fd int
)

func main() {
	parseTunIp()
	createTunInterface()
	setRoute()
	createRawSocket()

	go tcpListenerLoop()
	go tcpSendLoop()
	go tunReceiveLoop()

	<-make(chan interface{})
}

func parseTunIp() {
	var err error
	peerIp = net.ParseIP(os.Args[1]).To4()
	assert.AssertNotNil(peerIp, "peerIp invalid")

	peerPort, err = strconv.Atoi(os.Args[2])
	assert.AssertNil(err, "peerPort illegal")

	tunIp, tunNet, err = net.ParseCIDR(os.Args[3])
	assert.AssertNil(err, "network illegal")
	assert.AssertNotNil(tunIp, "network illegal")
	assert.AssertNotNil(tunNet, "network illegal")
	tunIp = tunIp.To4()

	log.Printf("tunIp='%s'", tunIp.String())
}

func createTunInterface() {
	var err error
	tunIf, err = water.New(water.Config{
		DeviceType: water.TUN,
	})
	assert.AssertNil(err, "failed to create tunIf")

	log.Printf("Tun Interface Name: %s\n", tunIf.Name())
}

func setRoute() {
	execCommand(fmt.Sprintf("ip address add %s dev %s", tunIp.String(), tunIf.Name()))

	execCommand(fmt.Sprintf("ip link set dev %s up", tunIf.Name()))

	execCommand(fmt.Sprintf("ip route add table main %s dev %s", tunNet.String(), tunIf.Name()))
}

func createRawSocket() {
	// create ip level raw socket
	var err error
	fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	assert.AssertNil(err, "failed to create raw socket")
}

func execCommand(command string) {
	log.Printf("exec command '%s'\n", command)

	cmd := exec.Command("/bin/bash", "-c", command)

	err := cmd.Run()
	assert.AssertNil(err, "failed to execute command")

	state := cmd.ProcessState
	assert.AssertTrue(state.Success(), fmt.Sprintf("exec command '%s' failed, code=%d", command, state.ExitCode()))
}

func tunReceiveLoop() {
	buffer := buf.NewByteBuffer(65536)
	packet := make([]byte, 65536)
	for {
		n, err := tunIf.Read(packet)

		assert.AssertNil(err, "failed to read data from tun")

		buffer.Write(packet[:n])
		for {
			frame, err := tunnel.ParseIPFrame(buffer)

			if err != nil {
				log.Println(err)
				buffer.Clean()
				break
			}
			if frame == nil {
				break
			}

			// transfer to peer side
			tcpPipe <- frame.ToBytes()

			log.Println("receive from tun, send through tunnel " + frame.String())
		}
	}
}

func tcpSendLoop() {
	var err error

	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", peerIp, peerPort))
	assert.AssertNil(err, "failed to parse tcpAddr")

	var conn *net.TCPConn

	log.Println("try to connect peer")

	conn, err = net.DialTCP("tcp", nil, tcpAddr)

	for {
		if err == nil {
			log.Println("connect peer success")
			break
		}

		log.Printf("try to reconnect 1s later, addr=%s, err=%v", tcpAddr.String(), err)

		time.Sleep(time.Second)

		conn, err = net.DialTCP("tcp", nil, tcpAddr)
	}

	for bytes := range tcpPipe {
		conn.Write(bytes)
	}
}

func tcpListenerLoop() {
	var err error

	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", "0.0.0.0", peerPort))
	assert.AssertNil(err, "failed to parse tcpAddr")

	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	assert.AssertNil(err, "failed to listener")

	log.Printf("listener on '%s'\n", tcpAddr.String())
	conn, err := tcpListener.AcceptTCP()
	assert.AssertNil(err, "failed to accept")

	log.Println("accept peer success")

	buffer := buf.NewByteBuffer(65536)
	packet := make([]byte, 65536)

	for {
		n, err := conn.Read(packet)
		assert.AssertNil(err, "failed to read from tcp tunnel")

		buffer.Write(packet[:n])

		for {
			frame, err := tunnel.ParseIPFrame(buffer)
			assert.AssertNil(err, "failed to parse ip package from tcp tunnel")

			if err != nil {
				log.Println(err)
				buffer.Clean()
				break
			}
			if frame == nil {
				break
			}

			log.Println("receive from tunnel, send through raw socket" + frame.String())

			// send ip frame through raw socket
			addr := syscall.SockaddrInet4{
				Addr: tunnel.IPToArray4(frame.Target),
			}
			err = syscall.Sendto(fd, frame.ToBytes(), 0, &addr)
			assert.AssertNil(err, "failed to send data through raw socket")
		}
	}
}
