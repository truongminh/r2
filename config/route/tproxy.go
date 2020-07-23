package route

import (
	"fmt"
	"r2/iptables"

	"github.com/vishvananda/netlink"
)

func SetupTProxy() error {
	err := ipForward()
	if err != nil {
		return err
	}
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	chains, err := ipt.ListChains("mangle")
	if err != nil {
		return err
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
	// add rule
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return err
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
			return fmt.Errorf("add rule mask: %s", err)
		}
	}
	return nil
}

func ClearTProxy() error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	ipt.ClearChain("mangle", "DIVERT")
	ipt.DeleteChain("mangle", "DIVERT")
	ipt.Delete("mangle", "PREROUTING", "-p", "tcp", "-m", "socket", "-j", "DIVERT")

	// clean rule
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if rule.Table == 100 && rule.Mark == 1 {
			err := netlink.RuleDel(&rule)
			if err != nil {
				return fmt.Errorf("delete rule mask: %s", err)
			}
		}
	}
	return nil
}
