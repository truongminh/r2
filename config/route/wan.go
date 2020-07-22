package route

import (
	"fmt"
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
