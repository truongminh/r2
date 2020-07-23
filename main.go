package main

import (
	"context"
	"fmt"
	"log"
	"r2/config"
	"r2/dns"
	"r2/proxy"
)

func main() {
	if config.V.Mode == "clear" {
		log.Printf("clear route")
		err := config.Clear()
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	config.Print()
	err := config.Setup()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listener, err := proxy.NewTCP(ctx, fmt.Sprintf(":%d", config.V.Port))
	if err != nil {
		log.Fatal(err)
	}
	listener.Handler = proxy.NewDelay(config.V.Delay, config.V.Hosts)
	// `lan := config.V.Lan
	// dbcpConfig := dhcp.ServerConfig{
	// 	Interface:  lan.Name,
	// 	ServerIP:   lan.Gateway,
	// 	RouterIP:   lan.Gateway,
	// 	SubnetMask: lan.Subnet.Mask,
	// 	StartIP:    config.V.DHCP.StartIP,
	// }
	// go dhcp.Start(dbcpConfig)
	go dns.Start(config.V.DNS.Addr)
	listener.Start()
}
