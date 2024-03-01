package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"
)

type InMemoryStore struct {
	data  map[string]string
	mutex sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
	}
}

func (s *InMemoryStore) Get(key string, result chan<- string) {
	go func() {
		s.mutex.RLock()
		defer s.mutex.RUnlock()
		value, ok := s.data[key]
		if !ok {
			result <- ""
			return
		}
		result <- value
	}()
}

// func (s *InMemoryStore) Get(key string, result chan<- string) {
// 	s.mutex.RLock()
// 	defer s.mutex.RUnlock()
// 	value, ok := s.data[key]
// 	if !ok {
// 		result <- ""
// 		return
// 	}
// 	result <- value
// }

func handleGet(store *InMemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		result := make(chan string)
		store.Get(key, result)
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
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.data[key] = value
	}()
}

func handlePut(store *InMemoryStore) http.HandlerFunc {
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

		store.Put(requestData.Key, requestData.Value)
		err = store.PersistDataToFile("data.json")
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
		s.mutex.Lock()
		defer s.mutex.Unlock()
		delete(s.data, key)
	}()
}

func handleDelete(store *InMemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		store.Delete(key)
		err := store.PersistDataToFile("data.json")
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

func (s *InMemoryStore) PersistDataToFile(filename string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(s.data); err != nil {
		return err
	}

	return nil
}

func loadDataFromFile(store *InMemoryStore, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&store.data)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	store := NewInMemoryStore()

	filename := "data.json"

	if _, err := os.Stat(""); errors.Is(err, os.ErrNotExist) {
		_, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Failed to create json file: %v", err)
		}
	}

	err := loadDataFromFile(store, "data.json")
	if err != nil {
		log.Fatalf("Failed to load data from file: %v", err)
	}

	http.HandleFunc("/get", handleGet(store))
	http.HandleFunc("/put", handlePut(store))
	http.HandleFunc("/delete", handleDelete(store))

	http.Handle("/", http.FileServer(http.Dir(".")))

	log.Printf("Server is running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
