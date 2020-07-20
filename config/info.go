package config

import (
	"log"
	"time"
)

func Print() {
	// log.Printf("%+v", V)
	log.Printf("port=%d, delay=%d ms", V.Port, V.Delay/time.Millisecond)
	log.Printf("hosts=%s", V.Hosts)
	log.Printf("wan=%+v lan=%+v", V.Wan, V.Lan)
}
