package p2p

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T){
	listenAddr := ":4000"
	tr := NewTCPTransport(listenAddr)

	assert.Equal(t,tr.ListenAddress, listenAddr, "ListenAddress should match the provided address")

  	err := tr.ListenAndAccept()
	assert.Nil(t, err, "ListenAndAccept should not return an error")

}  
