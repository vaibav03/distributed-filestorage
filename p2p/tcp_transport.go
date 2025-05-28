package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPTransport struct {
	ListenAddress string
	Listener      net.Listener
	peerLock		  sync.Mutex
	peers				 map[net.Addr]Peer
}

//represents remote node over a established TCP Connection
type TCPPeer struct{
	// conn -> underlying connection of peer
	conn net.Conn
	// dial 
	outbound bool
}

func NewTCPPeer(conn net.Conn , outbound bool) *TCPPeer {
		return &TCPPeer{
		conn : conn,
		outbound : outbound,
		}
}


func NewTCPTransport(listenAddr string) TCPTransport{
	return TCPTransport{
		ListenAddress: listenAddr,
	}
}

func (t * TCPTransport) ListenAndAccept()  error{
	var err error
	t.Listener, err = net.Listen("tcp", t.ListenAddress)
	if err!=nil{
		return err
	}

	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop(){
  for {
		conn, err := t.Listener.Accept()
		if err != nil {
		fmt.Println("TCP accepting connection error:", err)
		}

			go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true) 
	fmt.Println("Handling new connection from:", conn)

	fmt.Println("New peer connected:", peer)
	// Handle the connection (e.g., read/write messages)


}
