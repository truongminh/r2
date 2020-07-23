package dns

import (
	"log"

	"github.com/miekg/dns"
)

func Start(addr string) {
	proxy := &DNSProxy{
		DefaultServer: "8.8.8.8:53",
	}
	// attach request handler func
	dns.HandleFunc(".", proxy.Serve)
	server := &dns.Server{Addr: addr, Net: "udp"}
	log.Printf("Starting at %s\n", addr)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
