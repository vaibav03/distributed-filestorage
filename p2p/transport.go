package p2p

// Peer is an interface that defines the remote node.

type Peer interface{
 Close() error
}

// Transport is an interface that defines the methods for a transport layer in a peer-to-peer network.
// It allows for sending and receiving messages between peers. (TCP,UDP,WEBSOCKETS,....)
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
}

