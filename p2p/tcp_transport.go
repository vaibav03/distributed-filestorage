package p2p

import (
	"fmt"
	"net"
)

type TCPTransport struct {
	TCPTransportOpts
	listener      net.Listener
	peers				 map[net.Addr]Peer
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

type TCPTransportOpts struct{
	ListenAddress string
	HandshakeFunc func(any) error
	Decoder Decoder
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}

func (t * TCPTransport) ListenAndAccept()  error{
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err!=nil{
		return err
	}

	fmt.Printf("TCP Transport listening on %s\n", t.ListenAddress)

	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop(){
  for {
		conn, err := t.listener.Accept()

		if err != nil {
		fmt.Println("TCP accepting connection error:", err)
		}
		fmt.Println("TCP accepted connection from", conn.RemoteAddr())
			go t.handleConn(conn,false)
	}
}


func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	peer := NewTCPPeer(conn, true) 

	if err := t.HandshakeFunc(peer); err!=nil{
			fmt.Printf("Handshake failed for peer %s: %s\n", conn.RemoteAddr(), err)
		conn.Close()
		return 
	}

	buf := make([]byte, 1024) // Buffer size can be adjusted as needed


	for{
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Error reading from connection %s: %s\n", conn.RemoteAddr(), err)
			conn.Close()
			return
		}
		
		if n == 0 {
			fmt.Printf("Connection %s closed by peer\n", conn.RemoteAddr())
			conn.Close()
			return
		}
		fmt.Printf("Received message from %s: %v\n", conn.RemoteAddr(), buf[:n])
	}
}
