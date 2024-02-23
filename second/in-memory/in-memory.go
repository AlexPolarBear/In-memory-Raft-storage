package main

import (
	"encoding/json"
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

		// result := make(chan string)
		// go store.Get(key, result)
		// value := <-result

		// if value == "" {
		// 	http.Error(w, "Key not found", http.StatusNotFound)
		// 	return
		// }

		value, ok := store.Get(key)
		if !ok {
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

		res := "Successfully deleted key '" + key + "'"
		_, err := w.Write([]byte(res))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}

		r.Body.Close()
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
