package p2p 

// Message has data sent over each transport bw 2 peers
type RPC struct{
	From string
	Payload []byte
}