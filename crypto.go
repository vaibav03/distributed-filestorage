package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func newEncryptionKey() []byte {
	keyBuf := make([]byte, 32) // AES-256 requires a 32-byte key
	io.ReadFull(rand.Reader, keyBuf) // Fill the buffer with random bytes
	return keyBuf
}

func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int64, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}

	buf := make([]byte, 32*1024)
	stream := cipher.NewCTR(block, iv)
	nw := block.BlockSize()
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			
			 nn, err := dst.Write(buf[:n])
			 if err != nil {
				return 0, err
			}
			nw += nn
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
}

func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int64, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	iv := make([]byte,block.BlockSize())
	if _,err := io.ReadFull(rand.Reader,iv);err!=nil{
		return 0,err
	} 
	//prepend the IV to file 

	if _,err := dst.Write(iv);err!=nil{
		return 0,err
	}

	buf := make([]byte, 32*1024)
	stream := cipher.NewCTR(block,iv)
	nw := block.BlockSize() // to account for the IV size
	for{
		n,err :=  src.Read(buf)
		if n>0 {
			stream.XORKeyStream(buf, buf[:n])
			nn,err := dst.Write(buf[:n]);
			if err!=nil{
				return 0,err
			}
			nw+=nn
		}
		if err == io.EOF{
			break
		}
		if err!=nil{
			return 0,err
		}
	}
	return int64(nw),nil
}