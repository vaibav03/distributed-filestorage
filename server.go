package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	//	"time"

	"github.com/vaibav03/distributed-filestorage/p2p"
)
type Payload struct{
	Key string
	Data []byte
}

type Message struct{
	From string
	Payload any
}

func(s *FileServer) broadcast(msg *Message) error{
	fmt.Println("broadcast")
	peers := []io.Writer{}
	fmt.Println("peers in broadcast ",s.peers)
	for _,peer := range s.peers{
		peers = append(peers,peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) handleMessage(from string, msg *Message) error{
		return s.handleMessageStoreFile(from, *msg.Payload.(*MessageStoreFile))
}

func(s *FileServer) handleMessageStoreFile(from string,msg MessageStoreFile) error{
	fmt.Print("handleMessageStoreFile from ",from," msg ",msg.PathKey," size ",msg.Size,"\n")
	peer,ok := s.peers[from]
	if !ok{
		return fmt.Errorf("peer %s not found", from)
	}
	if err:= s.store.Write(msg.PathKey, io.LimitReader(peer,int64(msg.Size))); err != nil {
		return fmt.Errorf("error writing to store: %w", err)
	}
	peer.(*p2p.TCPPeer).Wg.Done() // signal that the stream is done
	return nil
}

type MessageStoreFile struct{
	PathKey string
	Size int64
}

func (s *FileServer) StoreData(key string, r io.Reader) error{
   
	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r,buf)

	// if err := s.store.Write(key,tee); err!=nil{
	// 	return err
	// }


	// p := &Payload{
	// 	Key:key,
	// 	Data : buf.Bytes(),
	// }

	// s.broadcast(&Message{
	// 	From:    "todo",
	// 	Payload: p.Data,
	// })
	// fmt.Println("store data")
	// return nil

	time.Sleep(2 * time.Second) // Simulate some processing delay
  buf := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{
			PathKey: key,
			Size : 10,
 		},
	}
	gob.Register(&MessageStoreFile{})

	if err := gob.NewEncoder((buf)).Encode(msg); err != nil {
		return fmt.Errorf("error encoding message: %w", err)
	}

	for _,peer := range s.peers {
		log.Printf("sending data to peer %s", peer)
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
		log.Printf("data sent to peer %s", peer.RemoteAddr())
	}
	  time.Sleep(time.Second)
		payload := []byte("THIS LARGE FILE")

		for _,peer := range s.peers {
		log.Printf("sending data to peer %s", peer)
		
		n,err :=  io.Copy(peer,bytes.NewReader(payload))
		if err !=nil{
			return err
		}
		fmt.Println("received and written bytes to disk: ",n)
	}
	return nil
}
type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         *p2p.TCPTransport
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
		log.Println("inserting in map  ",p.RemoteAddr())
	 s.peers[p.RemoteAddr().String()] = p
	 return nil
}


func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) loop(){
	fmt.Println("into loop fileserver")

	defer func(){
		log.Println("file server stopped user quit action")
		s.Transport.Close()
	}()
	
	for{
		select{
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil{
				fmt.Println("error decoding message: ", err)
			}

			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("error handling message: ", err)
				return 
			}

			
			// peer,ok := s.peers[rpc.From] 
			// if !ok{
			// 	log.Println("peer not found in map ",rpc.From)
			// }

			// b := make([]byte,1024)
			// if _,err := peer.Read(b); err != nil {
			// 	panic(err)
			// }
			// fmt.Printf("recv: %s",msg)

			// fmt.Printf("recv: %s",string(b))

			// peer.(*p2p.TCPPeer).Wg.Done() // signal that the stream is done

			// if err:= s.handleMessage(&p); err!=nil{
			// 	log.Println("error handling message: ", err)
			// }
			// fmt.Println(string(msg.Payload))
		case <-s.quitch:
		 return
		}
	}
}

func (s *FileServer) bootstrapNetwork() error{
	for _,addr := range s.BootstrapNodes{
		fmt.Println("bootstrapping  ",addr)
		 go func(addr string){
			fmt.Println("dialing ",addr)
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
