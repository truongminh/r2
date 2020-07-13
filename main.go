package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "http port")
	delay := flag.Int("delay", 0, "delay in milliseconds")
	hosts := flag.String("hosts", "", "list of hosts")
	flag.Parse()
	log.Printf("port=%d, delay=%d ms", *port, *delay)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listener, err := newTcp(ctx, fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	listener.qos = &qos{
		delay:      time.Millisecond * time.Duration(*delay),
		delayHosts: strings.Split(*hosts, ","),
	}
	listener.Serve()
}
