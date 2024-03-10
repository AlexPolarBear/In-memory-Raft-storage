package mvcc

import "time"

type Item struct {
	Value   string
	Version int64
}

type Storage struct {
	data map[string][]*Item
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string][]*Item),
	}
}

func (s *Storage) Get(key string) (*Item, error) {
	items, ok := s.data[key]
	if !ok || len(items) == 0 {
		return nil, nil
	}
	return items[len(items)-1], nil
}

func (s *Storage) Set(key string, value string) error {
	item := &Item{
		Value:   value,
		Version: time.Now().Unix(),
	}
	s.data[key] = append(s.data[key], item)
	return nil
}

func (s *Storage) GetVersion(key string, version int64) (*Item, error) {
	items, ok := s.data[key]
	if !ok {
		return nil, nil
	}
	for _, item := range items {
		if item.Version == version {
			return item, nil
		}
	}
	return nil, nil
}
