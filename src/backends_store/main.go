package main

import (
	"io/ioutil"
	"log"
	"net"

	"github.com/boltdb/bolt"
)

const (
	sockFile = "/lab/build/ngx_http_set_backend.sock"
	dbFile   = "/lab/build/backends.db"
)

func main() {
	l, err := net.Listen("unix", sockFile)
	if err != nil {
		log.Fatal(err)
	}

	//open database
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go processRequest(c, db)
	}
}

func processRequest(c net.Conn, db *bolt.DB) {
	log.Println("Received connection")
	defer c.Close()
	host, err := ioutil.ReadAll(c)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Received host:", string(host))
	c.Write([]byte("www.google.com"))
}
