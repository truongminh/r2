package main

import (
	"net"
	"syscall"
)

func tcpAddrToSocketAddr(addr *net.TCPAddr) syscall.Sockaddr {
	ipv4 := addr.IP.To4()
	ip := [4]byte{}
	copy(ip[:], ipv4)
	return &syscall.SockaddrInet4{Addr: ip, Port: addr.Port}
}

func getIP(addr net.Addr) string {
	return addr.(*net.TCPAddr).IP.String()
}
