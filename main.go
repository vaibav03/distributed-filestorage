package main

import (
	"log"

	"github.com/vaibav03/distributed-filestorage/p2p"
)
func OnPeer(peer p2p.Peer) error {
	peer.Close()
	return nil
}
func main(){
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		Decoder : p2p.NOPDecoder{},
		HandshakeFunc: p2p.NOPHandShakeFunc,
		OnPeer: OnPeer,
	}
	tr:= p2p.NewTCPTransport(tcpOpts)
	go func(){
		for{
			msg:= <- tr.Consume()
			log.Printf("Received message: %s", msg.Payload)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Error starting TCP transport: %v", err)
	}
	select {}
}