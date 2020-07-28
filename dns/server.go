package dns

import (
	"log"

	"github.com/miekg/dns"
)

func Start(addr string) {
	proxy := &DNSProxy{
		DefaultServer: "8.8.8.8:53",
		Cache:         map[string]*dns.RR{},
	}
	// attach request handler func
	dns.HandleFunc(".", proxy.Serve)
	log.Printf("dns server on %s\n", addr)
	go func() {
		server := &dns.Server{Addr: addr, Net: "tcp"}
		err := server.ListenAndServe()
		defer server.Shutdown()
		if err != nil {
			log.Fatalf("Failed to start server: %s\n ", err.Error())
		}
	}()
	go func() {
		server := &dns.Server{Addr: addr, Net: "udp"}
		err := server.ListenAndServe()
		defer server.Shutdown()
		if err != nil {
			log.Fatalf("Failed to start server: %s\n ", err.Error())
		}
	}()
}
