package main

import (
	"log"
	"github.com/vaibav03/distributed-filestorage/p2p"
)

func main(){
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		Decoder : p2p.GOBDecoder{},
		HandshakeFunc: p2p.NOPHandShakeFunc,
	}
	tr:= p2p.NewTCPTransport(tcpOpts)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Error starting TCP transport: %v", err)
	}
	select {}
}