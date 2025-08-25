package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

func generateID() string {
	idBuf := make([]byte, 32) 
	io.ReadFull(rand.Reader, idBuf)
	return hex.EncodeToString(idBuf)
}


func hashKey(key string) string {
	// Simple hash function for demonstration purposes
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

func newEncryptionKey() []byte {
	keyBuf := make([]byte, 32) // AES-256 requires a 32-byte key
	io.ReadFull(rand.Reader, keyBuf) // Fill the buffer with random bytes
	return keyBuf
}

func copyStream(stream cipher.Stream,blockSize int, src io.Reader, dst io.Writer) (int64,error){

	buf := make([]byte, 32*1024)
	nw := blockSize
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
	return int64(nw), nil
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

  stream := cipher.NewCTR(block, iv)

	return copyStream(stream, block.BlockSize(), src, dst)
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

	stream := cipher.NewCTR(block,iv)
	return copyStream(stream, block.BlockSize(), src, dst)
}