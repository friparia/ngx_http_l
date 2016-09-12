package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/boltdb/bolt"
)

const (
	socket  = "/lab/build/ngx_http_set_backend.sock"
	dbFile  = "/lab/build/backends.db"
	nobody  = "nobody"
	apiPort = "9999"
)

func main() {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		log.Fatal("$DOMAIN must be set")
	}

	//open database
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//start backend provider
	provider := &provider{
		socket:   socket,
		username: nobody,
		db:       db,
		internalMapping: map[string]string{
			fmt.Sprintf("api.%s", domain): "127.0.0.1:" + apiPort,
		},
	}
	defer provider.cleanup()
	go func() {
		if err := provider.listen(); err != nil {
			log.Fatal(err)
		}
	}()

	//start rest api
	go func() {
		if err := startApi(apiPort, db); err != nil {
			log.Fatal(err)
		}
	}()

	//signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}
