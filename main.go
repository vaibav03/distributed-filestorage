package main

import (
		"bytes"
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
		EncKey:            newEncryptionKey(),
		StorageRoot:       strings.Split(listenAddr, ":")[1] + "network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcptransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)

	
	tcptransport.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":3000")
	s2 := makeServer(":6000")
	s3 := makeServer(":5001", ":3000", ":6000")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()

	time.Sleep(2 * time.Second)

	go func() {	log.Fatal(s3.Start()) }()
	time.Sleep(2 * time.Second)

	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("my big data file here!"))
		s3.StoreData(key, data)

		if err := s3.store.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}

		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))
	}
	select{}
}