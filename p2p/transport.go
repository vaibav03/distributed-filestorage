package p2p

import "net"

// Peer is an interface that defines the remote node.
type Peer interface{
	net.Conn
	Send([]byte) error
}

// Transport is an interface that defines the methods for a transport layer in a peer-to-peer network.
// It allows for sending and receiving messages between peers. (TCP,UDP,WEBSOCKETS,....)
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}

