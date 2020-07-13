package main

import (
	"context"
	"net"
	"time"
)

type dns struct {
	resolver *net.Resolver
}

func newDnsResolver() *dns {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 30,
			}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
	return &dns{
		resolver: resolver,
	}
}

func (d *dns) resolve(ctx context.Context, hostname string) ([]string, error) {
	return net.DefaultResolver.LookupHost(ctx, hostname)
}
