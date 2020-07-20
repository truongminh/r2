package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"syscall"
)

type Handler interface {
	Serve(src net.Conn, dst net.Conn) error
}

type listener struct {
	net.Listener
	Handler
}

func listenControl(network, addr string, c syscall.RawConn) (err error) {
	var sockErr error
	err = c.Control(func(fd uintptr) {
		// allow all ip
		sockErr = syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
		if sockErr != nil {
			sockErr = fmt.Errorf("must have CAP_NET_ADMIN superuser privilege: %s", sockErr.Error())
		}
	})
	if err == nil {
		err = sockErr
	}
	return
}

// NewTCP proxy
func NewTCP(ctx context.Context, addr string) (li *listener, err error) {
	lc := net.ListenConfig{
		Control: listenControl,
	}
	ll, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return
	}
	log.Printf("tcp listen on %s", addr)
	li = &listener{
		Listener: ll,
	}
	return
}

func (li *listener) handle(conn net.Conn) {
	srcAddr := conn.RemoteAddr()
	log.Printf("accept tcp src=%s dst=%s", srcAddr.String(), conn.LocalAddr().String())
	defer conn.Close()
	dstAddr := conn.LocalAddr().(*net.TCPAddr)
	dst, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		log.Printf("dial tcp: %s", err)
		return
	}
	defer dst.Close()
	li.Handler.Serve(conn, dst)
}

func (li *listener) Start() {
	for {
		conn, err := li.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				log.Printf("temporary error %s", netErr)
				continue
			}
			log.Fatalf("accept error: %s", err)
			return
		}
		go li.handle(conn)
	}
}
