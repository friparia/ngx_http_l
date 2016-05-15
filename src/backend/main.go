package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/boltdb/bolt"
)

const (
	socket = "/lab/build/ngx_http_set_backend.sock"
	dbFile = "/lab/build/backends.db"
	nobody = "nobody"
)

func main() {

	backendProvider := &backendProvider{
		socket:   socket,
		username: nobody,
		dbFile:   dbFile,
	}
	defer backendProvider.cleanup()
	go func() {
		if err := backendProvider.listen(); err != nil {
			log.Fatal(err)
		}
	}()

	// signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

type backendProvider struct {
	socket   string
	username string
	listener net.Listener
	dbFile   string
	db       *bolt.DB
}

func (p *backendProvider) listen() error {
	//open database
	var err error
	p.db, err = bolt.Open(p.dbFile, 0600, nil)
	if err != nil {
		return err
	}

	//create unix socket
	os.RemoveAll(p.socket) //in case it was not destroyed properly on exit
	p.listener, err = net.Listen("unix", p.socket)
	if err != nil {
		return err
	}
	if err = chown(p.username, p.socket); err != nil {
		return err
	}

	//start listening
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			defer c.Close()
			host, err := bufio.NewReader(c).ReadString('\n')
			if err != nil {
				//TODO return defined web page
			}
			backend, err := p.lookupBackend(host)
			if _, err := c.Write([]byte(backend + "\n")); err != nil {
				//TODO handle error
			}
		}(conn)
	}
}

func (p *backendProvider) lookupBackend(host string) (string, error) {
	return "www.google.com", nil
}

func (p *backendProvider) cleanup() {
	p.listener.Close()
	os.RemoveAll(p.socket)
	p.db.Close()
}
