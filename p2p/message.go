package p2p 
import(
	"net"
)

// Message has data sent over each transport bw 2 peers
type Message struct{
	From net.Addr
	Payload []byte
}