package proxy

import (
	"bytes"
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

// NewDelay handler
func NewDelay(delay time.Duration, hosts []string) Handler {
	return &delayHandler{
		delay: delay,
		hosts: hosts,
	}
}

func (d *delayHandler) delayOf(host string) time.Duration {
	for _, s := range d.hosts {
		if strings.HasSuffix(host, s) {
			return d.delay
		}
	}
	return 0
}

type delayIO struct {
	read    int
	written int
	delay   time.Duration
	src     io.ReadWriter
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

func (d *delayIO) Read(buf []byte) (n int, err error) {
	n, err = d.src.Read(buf)
	if err != nil {
		return
	}
	d.read += n
	return
}

func (d *delayIO) Write(buf []byte) (n int, err error) {
	time.Sleep(d.delay)
	n, err = d.src.Write(buf)
	d.written += n
	return
}

func (d *delayHandler) Serve(src net.Conn, dst net.Conn) error {
	var handshakeBuf bytes.Buffer
	host, _, _ := extractSNI(io.TeeReader(src, &handshakeBuf))
	_, err := dst.Write(handshakeBuf.Bytes())
	if err != nil {
		return err
	}
	var rw io.ReadWriter = src
	if len(host) != 0 {
		// TLS 1.3 with SNI
		delay := d.delayOf(host)
		rw = &delayIO{src: src, delay: delay}
		log.Printf("host=%s delay=%dms", host, delay/time.Millisecond)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
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
