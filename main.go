package main

import (
	"context"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	listener, err := newTcp(ctx, ":8080")
	if err != nil {
		log.Fatal(err)
	}
	listener.Serve()
}
