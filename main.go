package main

import (
	"context"
	"fmt"
	"log"
	"r2/config"
	"r2/proxy"
)

func main() {
	if config.V.Mode == "clear" {
		log.Printf("clear route")
		config.Clear()
		return
	}
	config.Print()
	config.Setup()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listener, err := proxy.NewTCP(ctx, fmt.Sprintf(":%d", config.V.Port))
	if err != nil {
		log.Fatal(err)
	}
	listener.Handler = proxy.NewDelay(config.V.Delay, config.V.Hosts)
	listener.Start()
}
