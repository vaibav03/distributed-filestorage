package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type TCPTransportOpts struct{
	ListenAddress string
	HandshakeFunc Handshakefunc
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
	wg  *sync.WaitGroup
}

func (p *TCPPeer) CloseStream() {
	fmt.Println(">>> CloseStream called. Calling wg.Done()")
	p.wg.Done()
}

func NewTCPPeer(conn net.Conn , outbound bool) *TCPPeer {
		return &TCPPeer{
		conn : conn,
		outbound : outbound,
		wg : &sync.WaitGroup{},
		}
}

func (t  *TCPTransport) Addr() string{
	return t.ListenAddress
}
func( t *TCPTransport) Close() error{
	return t.listener.Close();
}

func (p *TCPPeer) Read(b []byte) (n int, err error) {
	return p.conn.Read(b)
}

func (p *TCPPeer) Write(b []byte) (n int, err error) {
	return p.conn.Write(b)
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

func (p *TCPPeer) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *TCPPeer) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

func (p *TCPPeer) SetDeadline(t time.Time) error {
	return p.conn.SetDeadline(t)
}


func (p *TCPPeer) SetReadDeadline(t time.Time) error {
	return p.conn.SetReadDeadline(t)
}


func (p *TCPPeer) SetWriteDeadline(t time.Time) error {
	return p.conn.SetWriteDeadline(t)
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
		fmt.Println("dropping peer conn: ",err)
		conn.Close()
	}()


	peer := NewTCPPeer(conn, outbound) 
	

	if err := t.HandshakeFunc(peer); err!=nil{
			fmt.Printf("Handshake failed for peer %s: %s\n", conn.RemoteAddr(), err)
		conn.Close()
		return 
	}

		fmt.Println("Handshake successful for peer ",t.OnPeer)
		if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			fmt.Printf("OnPeer callback failed for peer %s: %s\n", conn.RemoteAddr(), err)
			return 
		}
	}


	for{
		rpc := RPC{}
		 err := t.Decoder.Decode(conn,&rpc); 

		 if err!=nil{
			fmt.Println("TCP Read Error:", err)
			conn.Close()
			break
		 }
		  rpc.From = conn.RemoteAddr().String()

		 if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting... \n ", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf(" [%s] STream closed, resuming read stream, waiting ... \n", conn.RemoteAddr())
			continue 
		}
		 
		t.rpch <- rpc
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