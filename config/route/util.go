package route

import (
	"encoding/json"
	"log"

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
