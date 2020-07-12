package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
)

type tcpListener struct{ *net.TCPListener }

func newTcp(ctx context.Context, addr string) (listener *tcpListener, err error) {
	lc := net.ListenConfig{}
	network := "tcp"
	ll, err := lc.Listen(ctx, network, addr)
	if err != nil {
		return
	}
	li := ll.(*net.TCPListener)
	fd, err := li.File()
	if err != nil {
		err = &net.OpError{
			Op:   "listen",
			Net:  network,
			Addr: li.Addr(),
			Err:  fmt.Errorf("get file description %s", err.Error()),
		}
		return
	}
	defer fd.Close()
	err = syscall.SetsockoptInt(int(fd.Fd()), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
	if err != nil {
		err = &net.OpError{
			Op:   "listen",
			Net:  network,
			Addr: li.Addr(),
			Err:  fmt.Errorf("set sockopt %s", err.Error()),
		}
	}
	listener = &tcpListener{
		TCPListener: li,
	}
	return
}

func tcpAddrToSocketAddr(addr *net.TCPAddr) syscall.Sockaddr {
	ipv4 := addr.IP.To4()
	ip := [4]byte{}
	copy(ip[:], ipv4)
	return &syscall.SockaddrInet4{Addr: ip, Port: addr.Port}
}

func (listener *tcpListener) dial(dst *net.TCPAddr) (conn *net.TCPConn, err error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		err = &net.OpError{Op: "dial", Err: fmt.Errorf("socket open: %s", err)}
		return
	}
	// incoming IP_TRANSPARENT socket already used the port
	// reuse the port for the outgoing socket here
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		syscall.Close(fd)
		err = &net.OpError{Op: "dial", Err: fmt.Errorf("setsockopt SO_REUSEADDR: %s", err)}
		return
	}
	err = syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
	if err != nil {
		syscall.Close(fd)
		err = &net.OpError{Op: "dial", Err: fmt.Errorf("setsockopt IP_TRANSPARENT: %s", err)}
		return
	}
	// err = syscall.SetNonblock(fd, true)
	// if err != nil {
	// 	syscall.Close(fd)
	// 	err = &net.OpError{Op: "dial", Err: fmt.Errorf("setsockopt SO_NONBLOCK: %s", err)}
	// 	return
	// }
	err = syscall.Connect(fd, tcpAddrToSocketAddr(dst))
	if err != nil {
		if err2, ok := err.(syscall.Errno); ok && err2 == syscall.EINPROGRESS {
			// ignore error
		} else {
			syscall.Close(fd)
			err = &net.OpError{Op: "dial", Err: fmt.Errorf("socket connect: %s %s", err, dst.String())}
			return
		}
	}
	file := os.NewFile(uintptr(fd), "tcp-dial-"+dst.String())
	defer file.Close()
	netConn, err := net.FileConn(file)
	if err != nil {
		syscall.Close(fd)
		err = &net.OpError{Op: "dial", Err: fmt.Errorf("new net.Conn from fd: %s", err)}
		return
	}
	conn = netConn.(*net.TCPConn)
	return
}

func (listener *tcpListener) handle(conn *net.TCPConn) {
	log.Printf("accept tcp src=%s dst=%s", conn.RemoteAddr(), conn.LocalAddr().String())
	defer conn.Close()
	remoteAddr := conn.LocalAddr().(*net.TCPAddr)
	remote, err := listener.dial(remoteAddr)
	if err != nil {
		log.Printf("proxy connection: %s", err)
		return
	}
	defer remote.Close()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		_, err := io.Copy(remote, conn)
		log.Printf("done copy local to remote %s", err)
		wg.Done()
	}()
	go func() {
		_, err := io.Copy(conn, remote)
		log.Printf("done copy remote to local %s", err)
		wg.Done()
	}()
	wg.Wait()
}

func (listener *tcpListener) Serve() {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				log.Printf("temporary error %s", netErr)
				continue
			}
			log.Fatalf("accept error: %s", err)
			return
		}
		go listener.handle(conn)
	}
}
