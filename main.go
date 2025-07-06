package main

import (
//	"bytes"
	"fmt"
	 "io"
	"log"
	"strings"
	"time"

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
		StorageRoot:       strings.Split(listenAddr, ":")[1] + "network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcptransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)

	
	tcptransport.OnPeer = s.OnPeer

	return s
}

func main(){
	s1 := makeServer(":3000")
	s2 := makeServer(":4000",":3000")

	go func() {log.Fatal(s1.Start()) }()
	time.Sleep(2*time.Second)
	go s2.Start()
	time.Sleep(2*time.Second)

	
	// data := bytes.NewReader([]byte("my big data file"))
	// err := s2.StoreData("coolPicture.jpg",data)
	// if err!=nil{
	// 	log.Fatal(err)
	// } 

	r,err := s2.Get("coolPicture.jpg")
	if err!=nil{
		log.Fatal(err)
	}

	b,err := io.ReadAll(r)
	if err!=nil{
		log.Fatal(err)
	}
	fmt.Println("Printing from main ",string(b))

	select{}
}