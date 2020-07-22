package config

import (
	"net"
	"r2/config/route"
	"time"
)

type rawConfig struct {
	Port  int
	Delay int
	Hosts []string
	Wan   struct {
		Name    string
		Gateway string
	}
	Lan struct {
		Name    string
		Subnet  string
		Gateway string
		DHCP    []string
	}
	Mode string
}

type ParsedConfig struct {
	Port  int
	Delay time.Duration
	Hosts []string
	Wan   route.Network
	Lan   route.Network
	Lo    route.Network
	DHCP  struct {
		StartIP net.IP
	}
	Mode string
}

var V = &ParsedConfig{}
var raw = &rawConfig{}

func (c *ParsedConfig) apply(r *rawConfig) (err error) {
	c.Port = r.Port
	c.Delay = time.Duration(r.Delay) * time.Millisecond
	c.Hosts = r.Hosts
	c.Wan, err = route.NewNetwork(r.Wan.Name)
	if err != nil {
		return
	}
	c.Wan.Gateway = net.ParseIP(r.Wan.Gateway)
	c.Lan, err = route.NewNetwork(r.Lan.Name)
	if err != nil {
		return
	}
	c.Lan.Gateway = net.ParseIP(r.Lan.Gateway)
	_, c.Lan.Subnet, err = net.ParseCIDR(r.Lan.Subnet)
	if err != nil {
		return
	}
	c.Lo, err = route.NewNetwork("lo")
	if err != nil {
		return
	}
	c.Mode = r.Mode
	c.DHCP.StartIP = net.ParseIP(r.Lan.DHCP[0])
	return
}
