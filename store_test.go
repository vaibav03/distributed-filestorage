package main

import (
	"bytes"
	"io"
	"log"
	"testing"
)

 
func  newStore() *Store{
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}
func teardown(t *testing.T, s *Store){
	if err:= s.Clear(); err!=nil{
		t.Error(err)
	}
}

func TestStore(t *testing.T){
		opts := StoreOpts{
			PathTransformFunc: CASPathTransformFunc,
		}
		s:= NewStore(opts)
		defer teardown(t,s)
		key := "momsspecials"
		data := []byte("some jpg bytes")


		if err := s.WriteStream(key,bytes.NewReader(data)); err != nil{
			t.Error(err)
		}

		if ok:= s.Has(key); !ok{
			t.Errorf("expected to have key %s",key)
		}
	
		r,err := s.Read(key)
		if err!=nil{
			t.Error(err)
		}

		b,_ := io.ReadAll(r)

		log.Println(string(b))

		if string(b)!= string(data) {
			t.Errorf("Expected %v, got %v", data,b)
		}

		s.Delete(key)
		
}

func TestPath(t *testing.T){
	key := "momsbestpic"
	pathname := CASPathTransformFunc(key)

	log.Println(pathname)
}

func TestStoreDeleteKey(t *testing.T){
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s:= NewStore(opts)
	key := "momsspecials"
	data := []byte("some jpg bytes")

	if err:= s.WriteStream(key,bytes.NewReader(data)); err != nil{
		t.Error(err)
	}

	if err:= s.Delete(key); err != nil{
		t.Error(err)
	}
}