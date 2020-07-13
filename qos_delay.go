package main

import (
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type delayHandler struct {
	delay time.Duration
	hosts []string
}

type delayIO struct {
	io.Reader
	io.Writer
	count  int
	enable bool
	delay  time.Duration
	hosts  []string
}

func onlyReadable(buf []byte) string {
	s := []byte{}
	for _, c := range buf {
		if c > 32 && c < 127 {
			s = append(s, c)
		}
	}
	return string(s)
}

func (d *delayIO) Read(buf []byte) (int, error) {
	if d.count < 2 {
		s := onlyReadable(buf)
		if len(s) > 0 && strings.Contains(s, "h2http/") {
			for _, h := range d.hosts {
				if strings.Contains(s, h) {
					log.Printf("delay host=%s header=%s", h, s)
				}
				d.enable = true
			}
		}
		d.count++
	}
	return d.Reader.Read(buf)
}

func (d *delayIO) Write(buf []byte) (int, error) {
	if d.enable {
		time.Sleep(d.delay)
	}
	return d.Writer.Write(buf)
}

func (d *delayHandler) Serve(src net.Conn, dst net.Conn) error {
	wg := sync.WaitGroup{}
	wg.Add(2)
	rw := &delayIO{Reader: src, Writer: src, hosts: d.hosts, delay: d.delay}
	go func() {
		io.Copy(rw, dst)
		wg.Done()
	}()
	go func() {
		io.Copy(dst, rw)
		wg.Done()
	}()
	wg.Wait()
	return nil
}
