package p2p


type Handshakefunc func(Peer) error

func NOPHandShakeFunc(peer Peer) error {
return nil
}