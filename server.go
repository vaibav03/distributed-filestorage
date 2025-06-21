package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/vaibav03/distributed-filestorage/p2p"
)
type Payload struct{
	Key string
	Data []byte
}

func(s *FileServer) broadcast(p *Payload) error{
	fmt.Println("broadcast")
	peers := []io.Writer{}
	for _,peer := range s.peers{
		peers = append(peers,peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)
}
func (s *FileServer) StoreData(key string, r io.Reader) error{
  fmt.Println("store data")
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r,buf)

	if err := s.store.Write(key,tee); err!=nil{
		return err
	}


	fmt.Println(buf.Bytes())

	p := &Payload{
		Key:key,
		Data : buf.Bytes(),
	}

	return s.broadcast(p)
}
type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.TCPTransport
	BootstrapNodes 		[]string
}

type FileServer struct {
	FileServerOpts
	peerLock sync.Mutex
	peers map[string]p2p.Peer
	store *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store: NewStore(storeOpts),
		quitch : make(chan struct{}),
		peers : make(map[string]p2p.Peer),

	}
}

func ( s *FileServer) OnPeer(p p2p.Peer) error{
	 s.peerLock.Lock()
	 defer s.peerLock.Unlock()
	 s.peers[p.RemoteAddr().String()] = p
	 log.Println("connected with remote  ",p.RemoteAddr())
	 return nil
}


func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) loop(){
	fmt.Println("into loop")

	defer func(){
		log.Println("file server stopped user quit action")
		s.Transport.Close()
	}()
	
	for{
		select{
		case msg := <-s.Transport.Consume():
			var p Payload
			if err:=gob.NewDecoder((bytes.NewReader(msg.Payload))).Decode(&p); err!=nil{
				log.Fatal(err)
			}
			fmt.Printf("Payload ...   %+v\n",p)
			// Store the received data
			if err := s.store.Write(p.Key, bytes.NewReader(p.Data)); err != nil {
				log.Printf("failed to store received data: %v", err)
			}
		case <-s.quitch:
		 return
		}
	}
}

func (s *FileServer) bootstrapNetwork() error{
	for _,addr := range s.BootstrapNodes{
		fmt.Println("bootstrapping  ",addr)
		 go func(addr string){
			if err:= s.Transport.Dial(addr); err!=nil{
				log.Printf("dial error : %s ",err)
			}
		 }(addr)
	}
	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.bootstrapNetwork()
	s.loop()
	return nil
}
