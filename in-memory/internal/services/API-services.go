package services

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type InMemoryStore struct {
	Data           map[string]string
	Mutex          sync.RWMutex
	SnapCh         chan map[string]string
	LogFile        *os.File
	OperationLog   *OperationLog
	TransactionLog *TransactionLog
}

type OperationLog struct {
	Operations []string
}

type PathsConfig struct {
	DataFile string `json:"data_file"`
}

type TransactionLog struct {
	Transactions []LogEntry
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		Data:           make(map[string]string),
		SnapCh:         make(chan map[string]string),
		OperationLog:   &OperationLog{Operations: []string{}},
		TransactionLog: &TransactionLog{Transactions: []LogEntry{}},
	}
}

func (s *InMemoryStore) Get(key string) (string, bool) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	value, ok := s.Data[key]
	return value, ok
}

func (s *InMemoryStore) Put(key, value string) {
	go func() {
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		s.Data[key] = value
		s.LogOperation("PUT", key, value)
		if err := s.PersistLogToFile(); err != nil {
			log.Printf("Error persisting log: %v", err)
		}
	}()
}

func (s *InMemoryStore) Delete(key string) {
	go func() {
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		delete(s.Data, key)
		s.LogOperation("DELETE", key, "")
	}()
}

func (s *InMemoryStore) PersistDataToFile() error {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	file, err := os.Open("in-memory/configs/config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := PathsConfig{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.MarshalIndent(s.Data, "", "    ")
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = os.WriteFile(config.DataFile, data, 0644)
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = s.PersistLogToFile()
	if err != nil {
		return err
	}

	return nil
}

func PeriodicSave(s *InMemoryStore, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		err := s.PersistDataToFile()
		if err != nil {
			log.Fatal("Error saving data to file:", err)
		}
	}
}
