package dhcp

import (
	"log"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
)

// // DORAHandler is a server handler suitable for DORA transactions
// func handler(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
// 	if m == nil {
// 		log.Printf("Packet is nil!")
// 		return
// 	}
// 	if m.OpCode != dhcpv4.OpcodeBootRequest {
// 		log.Printf("Not a BootRequest!")
// 		return
// 	}
// 	reply, err := dhcpv4.NewReplyFromRequest(m)
// 	if err != nil {
// 		log.Printf("NewReplyFromRequest failed: %v", err)
// 		return
// 	}
// 	reply.UpdateOption(dhcpv4.OptServerIdentifier(net.IP{1, 2, 3, 4}))
// 	switch mt := m.MessageType(); mt {
// 	case dhcpv4.MessageTypeDiscover:
// 		reply.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeOffer))
// 	case dhcpv4.MessageTypeRequest:
// 		// reply.UpdateOption(dhcpv4.OptRouter)
// 		reply.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
// 	default:
// 		log.Printf("Unhandled message type: %v", mt)
// 		return
// 	}

// 	if _, err := conn.WriteTo(reply.ToBytes(), peer); err != nil {
// 		log.Printf("Cannot reply to client: %v", err)
// 	}
// 	log.Print(m.Summary())
// }

type ServerConfig struct {
	Interface  string
	ServerIP   net.IP
	SubnetMask net.IPMask
	StartIP    net.IP
	Length     byte
	Debug      bool
}

func Start(config ServerConfig) {
	if config.Length < 1 {
		config.Length = 8
	}
	laddr := &net.UDPAddr{
		Port: dhcpv4.ServerPort,
	}
	h := &Handler{
		ip:        config.ServerIP,
		leaseTime: time.Minute * 15,
		alloc:     newAllocMem(config.StartIP, config.Length),
		options: dhcpOptions{
			dhcpv4.OptionSubnetMask: dhcpv4.IP(config.SubnetMask),
			dhcpv4.OptionRouter:     dhcpv4.IP(config.ServerIP),
			dhcpv4.OptionDomainNameServer: dhcpv4.IPs([]net.IP{
				net.IP{8, 8, 8, 8},
			}),
		},
	}
	go func() {
		c := time.NewTimer(time.Minute)
		for {
			<-c.C
			before := time.Now().Add(-h.leaseTime)
			log.Printf("collect unused dhcp ip")
			h.alloc.Collect(before)
		}
	}()
	server, err := server4.NewServer(config.Interface, laddr, h.ServeDHCP)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
