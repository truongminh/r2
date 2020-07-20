package config

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
)

func loadFile() {
	configFile := "config.toml"
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("WARN: %s not exist", configFile)
			return
		}
		log.Fatal(err)
	}
	err = toml.Unmarshal(buf, V)
	if err != nil {
		log.Fatal(err)
	}
}

func parseFlags() {
	port := flag.Int("port", 0, "http port")
	delay := flag.Int("delay", 0, "delay in milliseconds")
	hosts := flag.String("hosts", "", "list of hosts")
	clear := flag.Bool("clear", false, "clear route")
	flag.Parse()
	if *port > 0 {
		V.Port = *port
	}
	if *delay > 0 {
		V.Delay = time.Duration(*delay)
	}
	if *clear {
		V.Mode = "clear"
	}
	if len(*hosts) > 0 {
		V.Hosts = strings.Split(*hosts, ",")
	}
}

func defaultConfig() {
	if V.Port == 0 {
		V.Port = 8080
	}
	// V.Wan.Name = "enp0s3"
	// V.Wan.Gateway = "10.0.2.2"
	// V.Lan.Name = "enp0s8"
	// V.Lan.Gateway = "192.168.10.1"
	// V.Lan.Subnet = "192.168.10.0/24"
}

func init() {
	loadFile()
	// parseFlags()
	defaultConfig()
	V.Delay *= time.Millisecond
}
