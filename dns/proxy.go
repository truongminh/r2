package dns

import (
	"fmt"
	"log"
	"regexp"

	"github.com/miekg/dns"
)

type DNSProxy struct {
	Cache         map[string]*dns.RR
	DefaultServer string
}

func (proxy *DNSProxy) getResponse(requestMsg *dns.Msg) (*dns.Msg, error) {
	responseMsg := new(dns.Msg)
	dnsServer := proxy.DefaultServer
	for _, question := range requestMsg.Question {
		switch question.Qtype {
		case dns.TypeA:
			answer, err := proxy.processTypeA(dnsServer, &question, requestMsg)
			if err != nil {
				return responseMsg, err
			}
			responseMsg.Answer = append(responseMsg.Answer, *answer)

		default:
			answer, err := proxy.processOtherTypes(dnsServer, &question, requestMsg)
			if err != nil {
				return responseMsg, err
			}
			responseMsg.Answer = append(responseMsg.Answer, *answer)
		}
	}

	return responseMsg, nil
}

func (proxy *DNSProxy) processOtherTypes(dnsServer string, q *dns.Question, requestMsg *dns.Msg) (*dns.RR, error) {
	queryMsg := new(dns.Msg)
	requestMsg.CopyTo(queryMsg)
	queryMsg.Question = []dns.Question{*q}

	msg, err := lookup(dnsServer, queryMsg)
	if err != nil {
		return nil, err
	}

	if len(msg.Answer) > 0 {
		return &msg.Answer[0], nil
	}
	return nil, fmt.Errorf("not found")
}

func (proxy *DNSProxy) processTypeA(dnsServer string, q *dns.Question, requestMsg *dns.Msg) (*dns.RR, error) {
	rr, ok := proxy.Cache[q.Name]
	if ok {
		return rr, nil
	}

	queryMsg := new(dns.Msg)
	requestMsg.CopyTo(queryMsg)
	queryMsg.Question = []dns.Question{*q}

	msg, err := lookup(dnsServer, queryMsg)
	if err != nil {
		return nil, err
	}

	if len(msg.Answer) > 0 {
		a := &msg.Answer[len(msg.Answer)-1]
		proxy.Cache[q.Name] = a
		return a, nil
	}
	return nil, fmt.Errorf("not found")
}

func lookup(server string, m *dns.Msg) (*dns.Msg, error) {
	dnsClient := new(dns.Client)
	dnsClient.Net = "tcp"
	response, _, err := dnsClient.Exchange(m, server)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (proxy *DNSProxy) Serve(w dns.ResponseWriter, r *dns.Msg) {
	switch r.Opcode {
	case dns.OpcodeQuery:
		m, err := proxy.getResponse(r)
		if err != nil {
			log.Printf("dns lookup %s with error: %s\n", r, err.Error())
			m.SetReply(r)
			w.WriteMsg(m)
			return
		}
		if len(m.Answer) > 0 {
			pattern := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
			ipAddress := pattern.FindAllString(m.Answer[0].String(), -1)

			if len(ipAddress) > 0 {
				log.Printf("Lookup for %s with ip %s\n", m.Answer[0].Header().Name, ipAddress[0])
			} else {
				log.Printf("Lookup for %s with response %s\n", m.Answer[0].Header().Name, m.Answer[0])
			}
		}
		m.SetReply(r)
		w.WriteMsg(m)
	}
}
