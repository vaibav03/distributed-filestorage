package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
)

type TCPTransportOpts struct{
	ListenAddress string
	HandshakeFunc func(any) error
	Decoder Decoder
	OnPeer func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener      net.Listener
	rpch chan RPC
}

//represents remote node over a established TCP Connection
// If your node initiated the connection, it's outbound, otherwise it's inbound.
type TCPPeer struct{
	conn net.Conn 
	outbound bool
}

func NewTCPPeer(conn net.Conn , outbound bool) *TCPPeer {
		return &TCPPeer{
		conn : conn,
		outbound : outbound,
		}
}

func( t *TCPTransport) Close() error{
	return t.listener.Close();
}


func(p *TCPPeer) Send(b []byte) error{
	_,err := p.conn.Write(b)
	return err
}
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpch : make(chan RPC),
	}
}

// return read only channel for reading incoming msgs received from peers
func (t *TCPTransport) Consume() <-chan RPC{
	return t.rpch
}
func (t * TCPTransport) ListenAndAccept()  error{
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err!=nil{
		return err
	}

	fmt.Printf("TCP Transport listening on %s\n", t.ListenAddress)

	go t.startAcceptLoop()
	log.Printf("TCP transport listening on port : %s",t.ListenAddress)
	return nil
}

func (t *TCPTransport) startAcceptLoop(){
  for {
		conn, err := t.listener.Accept()
		if errors.Is(err,net.ErrClosed) {
			return
		}
		
		if err != nil {
		fmt.Println("TCP accepting connection error:", err)
		}
		fmt.Println("TCP accepted connection from", conn.RemoteAddr())
			go t.handleConn(conn,false)
	}
}


func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error
	defer func(){
		fmt.Println("dropping peer conn: %s \n",err)
		conn.Close()
	}()


	peer := NewTCPPeer(conn, outbound) 

	if err := t.HandshakeFunc(peer); err!=nil{
			fmt.Printf("Handshake failed for peer %s: %s\n", conn.RemoteAddr(), err)
		conn.Close()
		return 
	}

	if t.OnPeer != nil {
		if err := t.OnPeer; err!=nil {
			return 
		}
	}

	rpc := &RPC{}

	for{
		 err := t.Decoder.Decode(conn,rpc); 
		 if err == net.ErrClosed{
			return 
		 }

		 if err!=nil{
			fmt.Println("TCP Read Error:", err)
			conn.Close()
		 }
		 rpc.From = conn.RemoteAddr()
		 t.rpch <- *rpc

		 fmt.Println("Received message from", conn.RemoteAddr(), ":", string(rpc.Payload))
	}
	
}


func (t *TCPTransport) Dial (addr string) error{
	conn,err := net.Dial("tcp",addr)
	if err!=nil{
		return err
	}

	go t.handleConn(conn,true)

	return nil
} 