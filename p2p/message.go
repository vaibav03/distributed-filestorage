package p2p

const IncomingStream = 0x2 
const IncomingMessage = 0x1
// Message has data sent over each transport bw 2 peers
type RPC struct{
	From string
	Payload []byte
	Stream bool
}