package config

import (
	"log"
)

func Print() {
	log.Printf("%+v", V)
	log.Printf("port=%d, delay=%d ms", raw.Port, raw.Delay)
	log.Printf("hosts=%s", V.Hosts)
	log.Printf("wan=%+v lan=%+v", V.Wan, V.Lan)
}
