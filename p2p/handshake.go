package p2p


type Handshakefunc func(any) error

func NOPHandShakeFunc(peer any) error {
return nil
}