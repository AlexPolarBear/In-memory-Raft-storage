package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func (s *InMemoryStore) Get(key string, result chan<- string) {
	go func() {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		value, ok := s.Data[key]
		if !ok {
			result <- ""
			return
		}
		result <- value
	}()
}

func HandleGet(s *InMemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		result := make(chan string)
		s.Get(key, result)
		value := <-result

		if value == "" {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}

		res := "Value for key '" + key + "': '" + value + "'"
		s.OperationLog.Operations = append(s.OperationLog.Operations,
			fmt.Sprintf("Get key '%s':'%s'", key, value))
		log.Printf("Get key '%s':'%s'", key, value)

		err := s.PersistDataToFile()
		if err != nil {
			http.Error(w, "Failed to persist data", http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(res))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}

		r.Body.Close()
	}
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

func HandlePut(s *InMemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestData struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		s.Put(requestData.Key, requestData.Value)

		res := "Successfully stored value for key '" + requestData.Key + "'"
		s.OperationLog.Operations = append(s.OperationLog.Operations,
			fmt.Sprintf("Put key '%s':'%s'", requestData.Key, requestData.Value))
		log.Printf("Put key '%s':'%s'", requestData.Key, requestData.Value)

		err = s.PersistDataToFile()
		if err != nil {
			http.Error(w, "Failed to persist data", http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(res))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}

		r.Body.Close()
	}
}

func (s *InMemoryStore) Delete(key string) {
	go func() {
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		delete(s.Data, key)
		s.LogOperation("DELETE", key, "")
	}()
}

func HandleDelete(s *InMemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		s.Delete(key)

		res := "Successfully deleted key '" + key + "'"
		s.OperationLog.Operations = append(s.OperationLog.Operations,
			fmt.Sprintf("Delete key '%s'", key))
		log.Printf("Delete key '%s'", key)

		err := s.PersistDataToFile()
		if err != nil {
			http.Error(w, "Failed to persist data", http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(res))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}

		r.Body.Close()
	}
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
