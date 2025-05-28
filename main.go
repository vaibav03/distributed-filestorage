package main

import (
	"log"

	"github.com/vaibav03/distributed-filestorage/p2p"
)
func main(){
	tr := p2p.NewTCPTransport(":3000")

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Error starting TCP transport: %v", err)
	}
	select {}
}