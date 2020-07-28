package dhcp

import (
	"log"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

type dhcpOptions map[dhcpv4.OptionCode]dhcpv4.OptionValue

// Handler is the dhpcd handler for serving requests.
type Handler struct {
	ip        net.IP
	options   dhcpOptions
	leaseTime time.Duration
	alloc     Allocator
}

func (h *Handler) configureReply(m *dhcpv4.DHCPv4, mt dhcpv4.MessageType) (*dhcpv4.DHCPv4, error) {
	rep, err := dhcpv4.NewReplyFromRequest(m)
	if err != nil {
		return nil, err
	}

	rep.UpdateOption(dhcpv4.OptMessageType(mt))
	rep.UpdateOption(dhcpv4.OptServerIdentifier(h.ip))
	rep.UpdateOption(dhcpv4.OptIPAddressLeaseTime(h.leaseTime))

	for opt, val := range h.options {
		rep.UpdateOption(dhcpv4.Option{Code: opt, Value: val})
	}

	return rep, nil
}

// ServeDHCP returns a dhcp response for a dhcp request.
func (h *Handler) ServeDHCP(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
	if m == nil {
		log.Printf("Packet is nil!")
		return
	}
	if m.OpCode != dhcpv4.OpcodeBootRequest {
		log.Printf("Not a BootRequest!")
		return
	}
	// log.Printf(m.Summary())
	switch m.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		log.Printf("received discover from %v", m.ClientHWAddr)
		ip, err := h.alloc.Allocate(m.ClientHWAddr, nil)
		if err != nil {
			log.Printf("Error allocating IP for %v: %v", m.ClientHWAddr, err)
			return
		}

		log.Printf("Generated lease for mac [%v] ip [%v]", m.ClientHWAddr, ip)

		rep, err := h.configureReply(m, dhcpv4.MessageTypeOffer)
		if err != nil {
			log.Printf("While configuring discover reply: %v", err)
			return
		}

		rep.YourIPAddr = ip

		if _, err := conn.WriteTo(rep.ToBytes(), peer); err != nil {
			log.Printf("Error replying to DHCP discover: %v", err)
			return
		}
	case dhcpv4.MessageTypeRequest:
		log.Printf("received request for %v from %v", m.ClientHWAddr, m.ClientHWAddr)

		preferredIP := net.IP(m.Options[uint8(dhcpv4.OptionRequestedIPAddress)])
		if preferredIP == nil {
			preferredIP = m.ClientIPAddr
		}

		ip, err := h.alloc.Allocate(m.ClientHWAddr, preferredIP)
		if err != nil {
			log.Printf("Error allocating IP for %v: %v", m.ClientHWAddr, err)
			// FIXME NAK here
			rep, err := h.configureReply(m, dhcpv4.MessageTypeNak)
			if err != nil {
				log.Printf("While configuring discover reply: %v", err)
				return
			}
			if _, err := conn.WriteTo(rep.ToBytes(), peer); err != nil {
				log.Printf("Error replying to DHCP request: %v", err)
				return
			}
			return
		}

		log.Printf("Lease obtained for mac [%v] ip [%v]", m.ClientHWAddr, ip)

		rep, err := h.configureReply(m, dhcpv4.MessageTypeAck)
		if err != nil {
			log.Printf("While configuring discover reply: %v", err)
			return
		}

		rep.YourIPAddr = ip

		if _, err := conn.WriteTo(rep.ToBytes(), peer); err != nil {
			log.Printf("Error replying to DHCP request: %v", err)
			return
		}
	case dhcpv4.MessageTypeRelease:
		h.alloc.Free(m.ClientHWAddr)
		log.Printf("received release")
	case dhcpv4.MessageTypeDecline:
		log.Printf("received decline")
	}
}
