package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "ggnetwork"

type PathKey struct{
	PathName string
	FileName string
}

func DefaultPathTransformFunc(key string) PathKey{
	return PathKey{
		PathName : key,
		FileName : key,
	} 
}
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLength := len(hashStr) / blockSize

	paths := make([]string,sliceLength)

	for i:=0; i<sliceLength;i++{
		from,to := i*blockSize,(i+1)*blockSize
		paths[i] = hashStr[from:to]
	}
	return PathKey{
		PathName : strings.Join(paths,"/"),
		FileName : hashStr,
	}
		
}
type PathTransformFunc func(string) PathKey

func (p PathKey) FullPath() string{
	return fmt.Sprintf("%s%s",p.PathName,p.FileName)
}


type StoreOpts struct {
	// root contains all folders/files of system
	Root string
	PathTransformFunc PathTransformFunc 
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil{
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if opts.Root == ""{
		opts.Root = defaultRootFolderName
	}

	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) WriteStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)
	pathNamewithRoot := s.Root + "/" + pathKey.PathName

	if err := os.MkdirAll(pathNamewithRoot,os.ModePerm); err != nil {
		return err
	}

	fullPath := pathKey.FullPath()
	fullPathWithRoot := s.Root + "/" + fullPath 

	f,err := os.Create(fullPathWithRoot)
	
	if err != nil {
		return err
	}
	defer f.Close()
	n,err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("Wrote %d bytes to %s", n, fullPathWithRoot)
	return nil
}

func (s *Store) Read(key string) (io.Reader,error) {
	f,err := s.readStream(key)
	if err != nil {
		return nil,err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_,err = io.Copy(buf,f)
	return buf,err 
	

}

func ( s *Store) readStream(key string) (io.ReadCloser,error){
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := s.Root + "/" + pathKey.FullPath()
  f,err := os.Open(fullPathWithRoot)

	if err!=nil {
		return nil,err
	}
	return f,nil
}

func (s PathKey) FirstPathName() string {
	paths := strings.Split(s.PathName, "/")
	if len(paths) ==0 {
		return ""
	} else {
		return paths[0]
	}
}


func (s *Store) Delete(key string) error{
	pathKey := s.PathTransformFunc(key)
	defer func(){
		log.Printf("deleted %s",pathKey.FullPath())
	}()

	return os.RemoveAll(pathKey.FirstPathName()); 
}


func (s *Store) Has( key string) bool{
	Pathkey := CASPathTransformFunc(key)
fullPathWithRoot := s.Root + "/" + Pathkey.FullPath()
	_,err := os.Stat(fullPathWithRoot)

	return !errors.Is(err,os.ErrNotExist)
}

func (s *Store) Clear () error{
	return os.RemoveAll(s.Root)
}

func (s *Store) Write(key string, r io.Reader) error{
	return s.WriteStream(key,r)
}