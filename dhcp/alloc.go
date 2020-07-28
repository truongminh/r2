package dhcp

import (
	"net"
	"time"

	"github.com/pkg/errors"
)

// ErrRangeExhausted is returned when the IP range is exhausted
var ErrRangeExhausted = errors.New("IP range exhausted")
var ErrAllocated = errors.New("IP allocated")

type Allocator interface {
	Allocate(mac net.HardwareAddr, preferred net.IP) (net.IP, error)
	Free(mac net.HardwareAddr) error
	Collect(before time.Time) error
}
