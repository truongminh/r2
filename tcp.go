package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
)

type tcpListener struct {
	*net.TCPListener
	*qos
}

func newTcp(ctx context.Context, addr string) (listener *tcpListener, err error) {
	lc := net.ListenConfig{}
	network := "tcp"
	ll, err := lc.Listen(ctx, network, addr)
	if err != nil {
		return
	}
	log.Printf("tcp listen on %s", addr)

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

func (listener *tcpListener) open() (fd int, err error) {
	fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		err = &net.OpError{Op: "dial", Err: fmt.Errorf("socket open: %s", err)}
		return
	}
	err = syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
	if err != nil {
		syscall.Close(fd)
		err = &net.OpError{Op: "dial", Err: fmt.Errorf("setsockopt IP_TRANSPARENT: %s", err)}
		return
	}
	// err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	// if err != nil {
	// 	syscall.Close(fd)
	// 	err = &net.OpError{Op: "dial", Err: fmt.Errorf("setsockopt SO_REUSEADDR: %s", err)}
	// 	return
	// }
	// err = syscall.SetNonblock(fd, true)
	// if err != nil {
	// 	syscall.Close(fd)
	// 	err = &net.OpError{Op: "dial", Err: fmt.Errorf("setsockopt SO_NONBLOCK: %s", err)}
	// 	return
	// }
	return
}

func (listener *tcpListener) dial(dst *net.TCPAddr) (conn *net.TCPConn, err error) {
	fd, err := listener.open()
	if err != nil {
		return
	}
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
	srcAddr := conn.RemoteAddr()
	log.Printf("accept tcp src=%s dst=%s", srcAddr.String(), conn.LocalAddr().String())
	defer conn.Close()
	dstAddr := conn.LocalAddr().(*net.TCPAddr)
	dst, err := listener.dial(dstAddr)
	if err != nil {
		log.Printf("proxy connection: %s", err)
		return
	}
	defer dst.Close()
	var h handler
	if listener.qos != nil {
		h = listener.qos.Handler("tcp", getIP(srcAddr), getIP(dstAddr))
	}
	if h == nil {
		h = &copyHandler{}
	}
	h.Serve(conn, dst)
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
