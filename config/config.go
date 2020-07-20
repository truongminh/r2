package config

import "time"

type config struct {
	Port  int
	Delay time.Duration
	Hosts []string
	Wan   struct {
		Name    string
		Gateway string
	}
	Lan struct {
		Name    string
		Subnet  string
		Gateway string
	}
	Mode string
}

var V = &config{}
