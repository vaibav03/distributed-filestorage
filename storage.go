package main

import (
	// "bytes"
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
	return fmt.Sprintf("%s/%s",p.PathName,p.FileName)
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

func (s *Store) WriteStream(id string ,key string, r io.Reader) (int64,error) {
	f,err := s.openFileForWriting(id,key)
	
	if err != nil {
		return 0,err
	}
	return io.Copy(f,r)
}

func (s *Store) Read(id string ,key string) (int64,io.Reader,error) {
	return s.readStream(id,key)
}

func ( s *Store) readStream(id string,key string) (int64,io.ReadCloser,error){
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := s.Root  + id + "/" + pathKey.FullPath()

  file,err := os.Open(fullPathWithRoot)

	if err!=nil{
		return 0,nil,err
	}

	fi,err := file.Stat()
	if err!=nil{
		return 0,nil,err
	}

	return fi.Size(),file,nil
}

func (s PathKey) FirstPathName() string {
	paths := strings.Split(s.PathName, "/")
	if len(paths) ==0 {
		return ""
	} else {
		return paths[0]
	}
}


func (s *Store) Delete(ID string, key string) error{
	pathKey := s.PathTransformFunc(key)
	defer func(){
		log.Printf("deleted %s",s.Root + ID + "/" + pathKey.FullPath())
	}()
	fullPathWithRoot := s.Root + ID + "/" + pathKey.FullPath()
	return os.RemoveAll(fullPathWithRoot); 
}


func (s *Store) Has(id string, key string) bool{
	Pathkey := s.PathTransformFunc(key)
fullPathWithRoot := s.Root + id +  "/" + Pathkey.FullPath()
	_,err := os.Stat(fullPathWithRoot)

	return !errors.Is(err,os.ErrNotExist)
}

func (s *Store) Clear () error{
	return os.RemoveAll(s.Root)
}

func (s *Store) Write(id string ,key string, r io.Reader) (int64,error){
	return s.WriteStream(id,key,r)
}

func (s *Store) WriteDecrypt (id string ,encKey []byte,key string, r io.Reader) (int64,error) {
	f,err := s.openFileForWriting(id,key)
	
	if err != nil {
		return 0,err
	}

	n,err := copyDecrypt(encKey,r,f)
	return int64(n),err
} 
 


func (s *Store) openFileForWriting(id string, key string) (*os.File,error){
		pathKey := s.PathTransformFunc(key)
	pathNamewithRoot := s.Root + id + "/" + pathKey.PathName
	fmt.Println("creating directories: ", pathNamewithRoot)

	if err := os.MkdirAll(pathNamewithRoot,os.ModePerm); err != nil {
		return nil,err
	}

	fullPath := pathKey.FullPath()
	fullPathWithRoot := s.Root + id + "/" + fullPath 

	fmt.Println("opening file for writing: ", fullPathWithRoot)

	return os.Create(fullPathWithRoot)
}