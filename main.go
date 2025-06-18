package main

import (
	"log"
	"github.com/vaibav03/distributed-filestorage/p2p"
)

func makeServer (listenAddr string, nodes ...string) *FileServer{
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddress : listenAddr,
		HandshakeFunc : p2p.NOPHandShakeFunc,
		Decoder : p2p.NOPDecoder{},
	}
	tcptransport := p2p.NewTCPTransport(tcptransportOpts)

  fileServerOpts := FileServerOpts{
		StorageRoot : listenAddr + "network",
		PathTransformFunc : CASPathTransformFunc,
		Transport : *tcptransport,
		BootstrapNodes: nodes,
	}
	return NewFileServer(fileServerOpts)
}

func main(){
	s1 := makeServer(":3000")
	s2 := makeServer(":4000",":3000")

	go func() { log.Fatal(s1.Start()) }()
	s2.Start()
}