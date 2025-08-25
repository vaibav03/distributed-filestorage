package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

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
type MessageGetFile struct {
	Key string

}

func (s *FileServer) Get(key string)  (io.Reader,error){
	 if s.store.Has(key) {
		fmt.Printf("[%s] serving file (%s) from local disk \n",s.Transport.Addr(),key)
		_,r,err := s.store.Read(key)
		return r,err
	 }

	 fmt.Printf("[%s] dont have file (%s) locally, fetching from network... \n",s.Transport.Addr(),key)
	 msg := Message{
		Payload : MessageGetFile{
			Key : key,
		},
	 }
	 gob.Register(MessageGetFile{})

	 if err := s.broadcast(&msg); err!=nil{
		return nil,err
	 }

	 time.Sleep(time.Millisecond * 500) 
	 for _,peer := range s.peers{
		var fileSize int64 
		binary.Read(peer,binary.LittleEndian,&fileSize) // read the size of the file
			fmt.Println("file size received from peer: ",fileSize)

	n, err := s.store.WriteDecrypt(s.EncKey,key,io.LimitReader(peer,fileSize))

	if err != nil {
			return nil, fmt.Errorf("error writing file to store: %w", err)
		}

		 fmt.Printf("[%s] received [%d] bytes over the network from (%s) \n ",s.Transport.Addr(),n,peer.RemoteAddr())
		 peer.CloseStream()
	 }
	 _,r,err :=  s.store.Read(key)
	 return r,err
}
func(s *FileServer) stream(msg *Message) error{
	peers := []io.Writer{}
	fmt.Println("peers in broadcast ",s.peers)
	for _,peer := range s.peers{
		peers = append(peers,peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast (msg *Message) error{
	 msgbuf := new(bytes.Buffer)
	 gob.Register(&MessageStoreFile{})

	if err := gob.NewEncoder((msgbuf)).Encode(msg); err != nil {
		return fmt.Errorf("error encoding message: %w", err)
	}
	for _,peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(msgbuf.Bytes()); err != nil {
			return err
		}
		log.Printf("data sent to peer %s", peer.RemoteAddr())
	}
	return nil
}
func (s *FileServer) handleMessage(from string, msg *Message) error{
	fmt.Printf("TYPE OF PAYLOAD : %T", msg.Payload	)
	switch v := msg.Payload.(type) {
	case *MessageStoreFile:
		{
			fmt.Println("handling message store file")
			return s.handleMessageStoreFile(from, *v)
		}
	case MessageGetFile:{
		fmt.Println("handling message get file")
		 return s.handleMessageGetFile(from,v)
	}
	}
		return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error{
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("(%s) need to serve file (%s) but it does not exist on disk",s.Transport.Addr(), msg.Key)
	} 
	fmt.Printf("[%s] serving file  (%s) over the network \n",s.Transport.Addr(), msg.Key)

	fileSize,r,err := s.store.Read(msg.Key)
	if err != nil {
		return fmt.Errorf("error reading file from store: %w", err)
	}

	 rc,ok :=  r.(io.ReadCloser)
	 if ok{
		fmt.Println("received a ReadCloser, closing it after use")
		defer rc.Close()
	 }

	peer,ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found", from)
	}
	peer.Send([]byte{p2p.IncomingStream}) // signal that this is a stream
	binary.Write(peer, binary.LittleEndian, fileSize) // send the size of the file
	n,err := io.Copy(peer,r)
	if err != nil {
		return fmt.Errorf("error sending file to peer: %w", err)
	}
	log.Printf("[%s] written %d bytes over the network to peer %s ",s.Transport.Addr(), n, from)
	return nil
}
func(s *FileServer) handleMessageStoreFile(from string,msg MessageStoreFile) error{
	peer,ok := s.peers[from]
	if !ok{
		return fmt.Errorf("peer %s not found", from)
	}
	if _,err:= s.store.Write(msg.PathKey, io.LimitReader(peer,int64(msg.Size))); err != nil {
		return fmt.Errorf("error writing to store: %w", err)
	}
	peer.CloseStream() // signal that the stream is done
	return nil
}

type MessageStoreFile struct{
	PathKey string
	Size int64
}

func (s *FileServer) StoreData(key string, r io.Reader) error{
	filebuf := new(bytes.Buffer)
	tee := io.TeeReader(r,filebuf)

	n,err := s.store.Write(key,tee); 
	if err!=nil{
		return err
	}
	msg := Message{
		Payload: &MessageStoreFile{
			PathKey: key,
			Size: n+16,
		},
	}

	 	if err := s.broadcast(&msg); err !=nil{
			return err
		}

		time.Sleep(time.Millisecond*5)

		peers := []io.Writer{}

		for _,peer := range s.peers{
			peers = append(peers,peer)
		}
	  mw := io.MultiWriter(peers...)
    mw.Write([]byte{p2p.IncomingStream})
		n,err = copyEncrypt(s.EncKey, filebuf, mw)
		
		if err!=nil{
			return err
		}
		// n,err :=  io.Copy(peer,filebuf)
		 fmt.Printf(" [%s] received and written (%d) bytes to disk: \n",s.Transport.Addr(),n)
	
	return nil
}
type FileServerOpts struct {
	EncKey []byte
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
