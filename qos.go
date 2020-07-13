package main

import (
	"io"
	"net"
	"sync"
	"time"
)

type handler interface {
	Serve(src net.Conn, dst net.Conn) error
}

type qos struct {
	delay      time.Duration
	delayHosts []string
}

func (q *qos) Handler(protocol string, srcIP string, dstIP string) handler {
	return &delayHandler{delay: q.delay, hosts: q.delayHosts}
}

type copyHandler struct{}

func (c *copyHandler) Serve(src net.Conn, dst net.Conn) error {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		io.Copy(src, dst)
		wg.Done()
	}()
	go func() {
		io.Copy(dst, src)
		wg.Done()
	}()
	wg.Wait()
	return nil
}
