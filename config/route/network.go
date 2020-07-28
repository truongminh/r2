package route

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

type Network struct {
	Name    string
	Subnet  *net.IPNet
	Gateway net.IP
	Index   int
	link    netlink.Link
}

func NewNetwork(name string) (network Network, err error) {
	network.Name = name
	network.link, err = netlink.LinkByName(name)
	if err != nil {
		err = fmt.Errorf("network interface %s: %s", name, err)
		return
	}
	network.Index = network.link.Attrs().Index
	return
}
