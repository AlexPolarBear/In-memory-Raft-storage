package services

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type PathsConfig struct {
	DataFile string `json:"data_file"`
}

type OperationLog struct {
	Operations []string
}

type InMemoryStore struct {
	Data         map[string]string
	Mutex        sync.RWMutex
	LogFile      *os.File
	OperationLog *OperationLog
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		Data:         make(map[string]string),
		OperationLog: &OperationLog{Operations: []string{}},
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
		_, err := w.Write([]byte(res))
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
		err = s.PersistDataToFile()
		if err != nil {
			http.Error(w, "Failed to persist data", http.StatusInternalServerError)
			return
		}

		res := "Successfully stored value for key '" + requestData.Key + "'"
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
		err := s.PersistDataToFile()
		if err != nil {
			http.Error(w, "Failed to persist data", http.StatusInternalServerError)
			return
		}

		res := "Successfully deleted key '" + key + "'"
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

	// err = s.PersistLogToFile()
	// if err != nil {
	// 	return err
	// }

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

// func (s *InMemoryStore) PersistLogToFile() error {
// 	logFile, err := os.Open("in-memory/internal/data/log.log")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer logFile.Close()

// 	logData, err := json.MarshalIndent(s.OperationLog, "", "    ")
// 	if err != nil {
// 		log.Fatal(err)
// 		return err
// 	}

// 	_, err = logFile.Write(logData)
// 	if err != nil {
// 		log.Fatal(err)
// 		return err
// 	}

// 	return nil
// }
