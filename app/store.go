package app

import (
	"encoding/gob"
	"os"
)

const StoragePath = "config/storage.bin"

type Store struct {
	data map[string]interface{}
}

func (s *Store) Write(key string, value interface{}) {
	s.data[key] = value
}

func (s *Store) Read(key string) interface{} {
	return s.data[key]
}

func (s *Store) save() error {
	file, err := os.OpenFile(StoragePath, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(file)
	return encoder.Encode(s.data)
}

func (s *Store) restore() error {
	s.data = make(map[string]interface{})

	file, err := os.OpenFile(StoragePath, os.O_RDONLY, 0400)
	if err != nil {
		return nil
	}

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&s.data)
	if err != nil {
		return nil
	}

	return nil
}