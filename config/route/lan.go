package route

import (
	"fmt"
	"r2/iptables"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func SetupLan(lan Network) error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	// PREROUTING -i enp0s8 -p tcp -j TPROXY --on-port 8080 --on-ip 0.0.0.0 --tproxy-mark 0x1/0x1
	ipt.AppendUnique("mangle", "PREROUTING", "-i", lan.Name, "-p", "tcp", "-j", "TPROXY", "--on-port", "8080", "--on-ip", "0.0.0.0", "--tproxy-mark", "0x1/0x1")

	addr := *lan.Subnet
	addr.IP = lan.Gateway
	// ip a add 192.168.10.1/24 dev enp0s8
	err = netlink.AddrAdd(lan.link, &netlink.Addr{IPNet: &addr})
	if err != nil && !isErrExist(err) {
		return fmt.Errorf("add lan gateway %s: %s", lan.Gateway, err)
	}
	// ip r add 192.168.10.0/24 via 192.168.10.1
	err = netlink.RouteAdd(&netlink.Route{
		LinkIndex: lan.Index,
		Scope:     netlink.SCOPE_LINK,
		Dst:       lan.Subnet,
		Src:       lan.Gateway,
		Protocol:  unix.RTPROT_KERNEL,
		Type:      unix.RTN_UNICAST,
		Table:     unix.RT_TABLE_MAIN,
	})
	if err != nil && !isErrExist(err) {
		return fmt.Errorf("lan add route %s: %s", lan.Gateway, err)
	}
	return nil
}

func ClearLan(lan Network) error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	ipt.Delete("mangle", "PREROUTING", "-i", lan.Name, "-p", "tcp", "-j", "TPROXY", "--on-port", "8080", "--on-ip", "0.0.0.0", "--tproxy-mark", "0x1/0x1")
	return err
}
