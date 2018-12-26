package kbs

import "errors"

type Storage struct {
	storage map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		storage: make(map[string]string),
	}
}

func (s *Storage) Get(key string) (string, error) {
	if v, ok := s.storage[key]; ok {
		return v, nil
	}
	return "", errors.New("Not Found")
}
func (s *Storage) Put(key, value string) {
	s.storage[key] = value
}
