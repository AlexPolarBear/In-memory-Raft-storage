package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func (s *InMemoryStore) Get(key string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, ok := s.data[key]
	return value, ok
}

func handleGet(store *InMemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		value, ok := store.Get(key)
		if !ok {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "Value for key '%s': %s", key, value)
	}
}

func (s *InMemoryStore) Put(key, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = value
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
		fmt.Fprintf(w, "Successfully stored value for key '%s'", requestData.Key)
	}
}

func (s *InMemoryStore) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.data, key)
}

func handleDelete(store *InMemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		store.Delete(key)
		fmt.Fprintf(w, "Successfully deleted key '%s'", key)
	}
}

func main() {
	store := NewInMemoryStore()

	http.HandleFunc("/get", handleGet(store))
	http.HandleFunc("/put", handlePut(store))
	http.HandleFunc("/delete", handleDelete(store))

	http.Handle("/", http.FileServer(http.Dir(".")))

	log.Printf("Server is running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
