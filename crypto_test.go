package main

import (
	"bytes"
	"fmt"
	"testing"
)


func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "Foo not bar"
	// Test the copyEncrypt function with a sample key and data
	src := bytes.NewReader([]byte("Foo not Bar"))
	dst := new(bytes.Buffer)
	key := newEncryptionKey() // Generate a new encryption key

	_, err := copyEncrypt(key, src,dst)
	if err != nil {
		t.Fatalf("copyEncrypt failed: %v", err)
	}

	out := new(bytes.Buffer)
	nw,err := copyDecrypt(key,dst,out);
	if err!=nil{
		t.Error(err)
	}

	if nw != int64(16+len(payload)) {
		t.Errorf("Expected %d bytes, got %d bytes", 16+len(payload), nw)
	}

	if out.String()!= payload{
		t.Errorf("Expected %q, got %q", payload, out.String())
	}
	fmt.Println(out.String())
}