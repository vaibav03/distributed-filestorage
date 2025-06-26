package p2p
import (
	"io"
	"encoding/gob"
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
	buf := make([]byte,1024)

	n,err := r.Read(buf)

	if err!=nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}