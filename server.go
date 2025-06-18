package main

import (
	"fmt"
	"log"

	"github.com/vaibav03/distributed-filestorage/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.TCPTransport
	BootstrapNodes 		[]string
}

type FileServer struct {
	FileServerOpts
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
	}
}


func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) loop(){
	defer func(){
		log.Println("file server stopped user quit action")
	}()
	
	for{
		select{
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
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
