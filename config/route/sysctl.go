package route

import (
	"io/ioutil"
	"os"
)

func sysctlSet(key string, value string) error {
	return ioutil.WriteFile("/proc/sys/net/"+key, []byte(value), os.ModeAppend)
}

func ipForward() error {
	// echo 1 > /proc/sys/net/ipv4/ip_forward
	err := sysctlSet("ipv4/ip_forward", "1")
	if err != nil {
		return err
	}
	err = sysctlSet("ipv6/conf/all/forwarding", "1")
	return err
}
