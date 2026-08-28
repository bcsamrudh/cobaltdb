package main

import (
	"flag"
	"log"

	"github.com/bcsamrudh/cobaltdb/internal/server"
	"github.com/bcsamrudh/cobaltdb/internal/store"
)

func main() {
	address := flag.String("addr", "127.0.0.1:6380", "TCP address to listen on")
	flag.Parse()

	database := server.New(store.New())
	log.Printf("CobaltDB listening on %s", *address)
	if err := database.ListenAndServe(*address); err != nil {
		log.Fatal(err)
	}
}
