package config

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"os/exec"
	"r2/iptables"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/vishvananda/netlink"
)

func contain(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

func runCmd(argsList string) (stdout string, stderr string, err error) {
	log.Printf(argsList)
	args := strings.Split(argsList, " ")
	log.Printf(argsList)
	buf := &bytes.Buffer{}
	cmd := exec.Cmd{
		Args:   args[1:],
		Stdout: nil,
		Stderr: buf,
		Path:   args[0],
	}
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	stderr = buf.String()
	if len(stderr) > 0 {
		log.Printf(stderr)
	}
	return
}

func runMany(list ...string) {
	for _, arg := range list {
		runCmd(arg)
	}
}

func Clear() {
	ipt, err := iptables.New()
	if err != nil {
		log.Fatal(err)
	}
	ipt.ClearChain("mangle", "DIVERT")
	ipt.DeleteChain("mangle", "DIVERT")
	ipt.Delete("mangle", "PREROUTING", "-p", "tcp", "-m", "socket", "-j", "DIVERT")
	ipt.Delete("mangle", "PREROUTING", "-i", V.Lan.Name, "-p", "tcp", "-j", "TPROXY", "--on-port", "8080", "--on-ip", "0.0.0.0", "--tproxy-mark", "0x1/0x1")
	ipt.Delete("nat", "POSTROUTING", "-o", V.Wan.Name, "-j", "MASQUERADE")

	// clean rule
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		log.Fatal(err)
	}
	for _, rule := range rules {
		if rule.Table == 100 && rule.Mark == 1 {
			err := netlink.RuleDel(&rule)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// clean route
	links, err := netlink.LinkList()
	if err != nil {
		log.Fatal(err)
	}
	var lo netlink.Link
	for _, li := range links {
		if li.Attrs().Name == "lo" {
			lo = li
		}
	}
	err = netlink.RouteDel(&netlink.Route{
		LinkIndex: lo.Attrs().Index,
		Dst:       &net.IPNet{IP: net.IPv4(0, 0, 0, 0), Mask: net.IPv4Mask(0, 0, 0, 0)},
		Scope:     unix.RT_SCOPE_HOST,
		Type:      unix.RTN_LOCAL,
		Table:     100,
	})
	if err != nil {
		if err.Error() == "file exists" {
			log.Printf("lo all exists")
		} else {
			log.Fatal(err)
		}
	}
}

func Setup() {
	ipt, err := iptables.New()
	if err != nil {
		log.Fatal(err)
	}
	chains, err := ipt.ListChains("mangle")
	if err != nil {
		log.Fatal(err)
	}
	if !contain(chains, "DIVERT") {
		// iptables -t mangle -N DIVERT
		ipt.NewChain("mangle", "DIVERT")
	}
	// iptables -t mangle -A PREROUTING -p tcp -m socket -j DIVERT
	ipt.AppendUnique("mangle", "PREROUTING", "-p", "tcp", "-m", "socket", "-j", "DIVERT")
	// iptables -t mangle -A DIVERT -j MARK --set-mark 1
	ipt.AppendUnique("mangle", "DIVERT", "-j", "MARK", "--set-mark", "1")
	// iptables -t mangle -A DIVERT -j ACCEPT
	ipt.AppendUnique("mangle", "DIVERT", "-j", "ACCEPT")
	// PREROUTING -i enp0s8 -p tcp -j TPROXY --on-port 8080 --on-ip 0.0.0.0 --tproxy-mark 0x1/0x1
	ipt.AppendUnique("mangle", "PREROUTING", "-i", V.Lan.Name, "-p", "tcp", "-j", "TPROXY", "--on-port", "8080", "--on-ip", "0.0.0.0", "--tproxy-mark", "0x1/0x1")
	// iptables -t nat -A POSTROUTING -o enp0s3 -j MASQUERADE
	ipt.AppendUnique("nat", "POSTROUTING", "-o", V.Wan.Name, "-j", "MASQUERADE")

	// add rule
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		log.Fatal(err)
	}
	existed := false
	for _, li := range rules {
		if li.Table == 100 && li.Mark == 1 {
			existed = true
		}
	}
	if !existed {
		err = netlink.RuleAdd(&netlink.Rule{
			Priority:          32765,
			Mark:              1,
			Mask:              4294967295,
			Table:             100,
			Goto:              -1,
			Flow:              -1,
			SuppressIfgroup:   -1,
			SuppressPrefixlen: -1,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	// add route
	links, err := netlink.LinkList()
	if err != nil {
		log.Fatal(err)
	}
	var wan, lan, lo netlink.Link
	for _, li := range links {
		if li.Attrs().Name == V.Lan.Name {
			lan = li
		} else if li.Attrs().Name == V.Wan.Name {
			wan = li
		} else if li.Attrs().Name == "lo" {
			lo = li
		}
	}
	// ip r add default via 10.0.2.2 dev enp0s3
	err = netlink.RouteAdd(&netlink.Route{
		LinkIndex: wan.Attrs().Index,
		Gw:        net.ParseIP(V.Wan.Gateway),
		Protocol:  unix.RTPROT_BOOT,
		Type:      unix.RTN_UNICAST,
		Table:     unix.RT_TABLE_MAIN,
	})
	if err != nil {
		if err.Error() == "file exists" {
			// log.Printf("default gateway exists")
		} else {
			log.Fatalf("wan gateway %s: %s", V.Wan.Gateway, err)
		}
	}
	// ip r add 192.168.10.0/24 via 192.168.10.1
	_, dst, err := net.ParseCIDR(V.Lan.Subnet)
	if err != nil {
		log.Fatalf("lan subnet %s: %s", V.Lan.Subnet, err)
	}
	err = netlink.RouteAdd(&netlink.Route{
		LinkIndex: lan.Attrs().Index,
		Scope:     netlink.SCOPE_LINK,
		Dst:       dst,
		Src:       net.ParseIP(V.Lan.Gateway),
		Protocol:  unix.RTPROT_KERNEL,
		Type:      unix.RTN_UNICAST,
		Table:     unix.RT_TABLE_MAIN,
	})
	if err != nil {
		if err.Error() == "file exists" {
			// log.Printf("lan gateway exists")
		} else {
			log.Fatalf("lan %s: %s", V.Lan.Gateway, err)
		}
	}
	//  ip route add local 0.0.0.0/0 dev lo table 100
	err = netlink.RouteAdd(&netlink.Route{
		LinkIndex: lo.Attrs().Index,
		Dst:       &net.IPNet{IP: net.IPv4(0, 0, 0, 0), Mask: net.IPv4Mask(0, 0, 0, 0)},
		Scope:     unix.RT_SCOPE_HOST,
		Type:      unix.RTN_LOCAL,
		Table:     100,
	})
	if err != nil {
		if err.Error() == "file exists" {
			// log.Printf("lo all exists")
		} else {
			log.Fatal(err)
		}
	}
}

func listRoutes() {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		log.Fatal(err)
	}
	for _, li := range routes {
		if li.LinkIndex == 1 {
			buf, _ := json.Marshal(li)
			log.Printf("%s", buf)
		}
	}
}
