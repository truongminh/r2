package dhcp

import (
	"log"
	"net"
	"sync"
	"time"
)

type lease struct {
	offset byte
	ip     net.IP
	mac    net.HardwareAddr
	mtime  time.Time
}

type allocMem struct {
	start  net.IP
	length byte
	used   []byte
	sync.Mutex
	leases map[string]*lease
}

func newAllocMem(start net.IP, length byte) Allocator {
	return &allocMem{
		start:  start,
		length: length,
		used:   make([]byte, length),
		leases: map[string]*lease{},
	}
}

// already an IP present in the leases table for this mac, to renew the lease
// if necessary.
func (a *allocMem) Allocate(mac net.HardwareAddr, preferred net.IP) (net.IP, error) {
	a.Lock()
	defer a.Unlock()
	le, ok := a.leases[mac.String()]
	if ok {
		le.mtime = time.Now()
		return le.ip, nil
	}
	if preferred != nil {
		offset := preferred[len(preferred)-1] - a.start[len(a.start)-1]
		if a.used[offset] != 0 {
			return nil, ErrAllocated
		}
	}
	free := -1
	for i, used := range a.used {
		if used == 0 {
			free = i
			break
		}
	}
	if free == -1 {
		return nil, ErrRangeExhausted
	}
	off := byte(free)
	ip := make(net.IP, len(a.start))
	copy(ip, a.start)
	ip[len(ip)-1] += off
	le = &lease{
		offset: off,
		ip:     ip,
		mac:    mac,
		mtime:  time.Now(),
	}
	a.add(le)
	log.Printf("dhcp allocated mac=%s ip=%s offset=%d", mac, ip, free)
	return ip, nil
}

func (a *allocMem) Free(mac net.HardwareAddr) error {
	a.Lock()
	defer a.Unlock()
	le, ok := a.leases[mac.String()]
	if ok {
		a.del(le)
	}
	return nil
}

func (a *allocMem) add(le *lease) error {
	a.leases[le.mac.String()] = le
	a.used[le.offset] = 1
	return nil
}

func (a *allocMem) del(le *lease) error {
	delete(a.leases, le.mac.String())
	a.used[le.offset] = 0
	return nil
}

func (a *allocMem) Collect(mtime time.Time) error {
	a.Lock()
	defer a.Unlock()
	for _, le := range a.leases {
		if le.mtime.Before(mtime) {
			a.del(le)
		}
	}
	return nil
}
