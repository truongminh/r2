package route

import (
	"encoding/json"
	"log"
	"net"

	"github.com/vishvananda/netlink"
)

func isErrExist(err error) bool {
	return err.Error() == "file exists"
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

func contain(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

func getOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}
