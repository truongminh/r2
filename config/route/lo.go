package route

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func SetupLo(lo Network) error {
	//  ip route add local 0.0.0.0/0 dev lo table 100
	err := netlink.RouteAdd(&netlink.Route{
		LinkIndex: lo.Index,
		Dst:       &net.IPNet{IP: net.IPv4(0, 0, 0, 0), Mask: net.IPv4Mask(0, 0, 0, 0)},
		Scope:     unix.RT_SCOPE_HOST,
		Type:      unix.RTN_LOCAL,
		Table:     100,
	})
	if err != nil && !isErrExist(err) {
		return fmt.Errorf("setup lo: %s", err)
	}
	return nil
}

func ClearLo(lo Network) error {
	err := netlink.RouteDel(&netlink.Route{
		LinkIndex: lo.Index,
		Dst:       &net.IPNet{IP: net.IPv4(0, 0, 0, 0), Mask: net.IPv4Mask(0, 0, 0, 0)},
		Scope:     unix.RT_SCOPE_HOST,
		Type:      unix.RTN_LOCAL,
		Table:     100,
	})
	if err != nil {
		return fmt.Errorf("clear lo route: %s", err)
	}
	return err
}
