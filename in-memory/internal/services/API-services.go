package services

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
)

type InMemoryStore struct {
	Data  map[string]string
	Mutex sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		Data: make(map[string]string),
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

func HandleGet(store *InMemoryStore) http.HandlerFunc {
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
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		s.Data[key] = value
	}()
}

func HandlePut(store *InMemoryStore) http.HandlerFunc {
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
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		delete(s.Data, key)
	}()
}

func HandleDelete(store *InMemoryStore) http.HandlerFunc {
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
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(s.Data); err != nil {
		return err
	}

	return nil
}
