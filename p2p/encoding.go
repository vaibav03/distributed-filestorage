package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader,*RPC) error
}

type GOBDecoder struct{

}

func (dec  GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

type NOPDecoder struct{
}

func (dec NOPDecoder) Decode(r io.Reader, msg *RPC) error{
	peekBuf := make([]byte,1)

	if  _,err:=r.Read(peekBuf);err!=nil{
		return err
	}
	stream := peekBuf[0] == IncomingStream

	// Incase of stream no decoding what is being ent over network , We are just setting stream true,
	if stream {
		msg.Stream = true
		return nil
	}



	buf := make([]byte,1024)

	n,err := r.Read(buf)

	if err!=nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}