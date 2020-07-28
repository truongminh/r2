package route

import (
	"fmt"
	"log"
	"r2/iptables"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func SetupWan(wan Network) error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	// iptables -t nat -A POSTROUTING -o enp0s3 -j MASQUERADE
	err = ipt.AppendUnique("nat", "POSTROUTING", "-o", wan.Name, "-j", "MASQUERADE")
	if err != nil {
		return fmt.Errorf("iptables -t nat -A POSTROUTING -o enp0s3 -j MASQUERADE: %s", err)
	}
	routes, err := netlink.RouteList(wan.link, netlink.FAMILY_ALL)
	if err != nil {
		return err
	}
	if wan.Gateway == nil {
		for _, route := range routes {
			if route.Gw != nil {
				// default
				wan.Gateway = route.Gw
				log.Printf("detected wan gateway %s", route.Gw)
				break
			}
		}
	}
	if wan.Gateway == nil {
		return fmt.Errorf("wan interface %s without gateway", wan.Name)
	}
	// add route
	// ip r add default via 10.0.2.2 dev enp0s3
	err = netlink.RouteAdd(&netlink.Route{
		LinkIndex: wan.Index,
		Gw:        wan.Gateway,
		Protocol:  unix.RTPROT_BOOT,
		Type:      unix.RTN_UNICAST,
		Table:     unix.RT_TABLE_MAIN,
	})
	if err != nil && !isErrExist(err) {
		err = fmt.Errorf("wan gateway %s: %s", wan.Gateway, err)
		return err
	}
	return nil
}

func ClearWan(wan Network) error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	ipt.Delete("nat", "POSTROUTING", "-o", wan.Name, "-j", "MASQUERADE")
	return err
}

func init() {
	ip, err := getOutboundIP()
	if err != nil {
		log.Fatal(err)
	}
	addrs, err := netlink.AddrList(nil, netlink.FAMILY_ALL)
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range addrs {
		if a.IP.Equal(ip) {
			log.Printf("%+v", a)
		}
	}
}
