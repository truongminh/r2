package config

import (
	"log"
)

var version = "0.0.1"

func Print() {
	// log.Printf("%+v", V)
	log.Printf("R2 version %s", version)
	log.Printf("port=%d, delay=%d ms", raw.Port, raw.Delay)
	log.Printf("hosts=%s", V.Hosts)
	log.Printf("wan=%+v lan=%+v", V.Wan, V.Lan)
}
