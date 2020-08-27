package app

import (
	"encoding/gob"
	"log"
	"os"
	"reflect"
)

type Store struct {
	filepath string
	links map[string]interface{}
	empties map[string]interface{}
}

func (s *Store) Link(key string, link interface{}, empty interface{}) {
	if s.links == nil {
		s.links = make(map[string]interface{})
	}

	if s.empties == nil {
		s.empties = make(map[string]interface{})
	}

	v := reflect.Indirect(reflect.ValueOf(link))
	if !v.CanSet() {
		log.Println("Link destination for", key, "must be settable")
		return
	}

	gob.Register(v.Interface())

	s.links[key] = link
	s.empties[key] = empty
}

func (s *Store) save() error {
	file, err := os.Create(s.filepath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	data := make(map[string]interface{})
	for k, v := range s.links {
		data[k] = v
	}

	encoder := gob.NewEncoder(file)
	return encoder.Encode(data)
}

func (s *Store) restore() error {
	data := make(map[string]interface{})

	file, err := os.Open(s.filepath)
	if err != nil {
		return nil
	}
	defer func() {
		_ = file.Close()
	}()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	for k := range s.links {
		reflect.Indirect(reflect.ValueOf(s.links[k])).Set(reflect.ValueOf(s.empties[k]))
	}

	for k, v := range data {
		reflect.Indirect(reflect.ValueOf(s.links[k])).Set(reflect.Indirect(reflect.ValueOf(v)))
	}

	return nil
}