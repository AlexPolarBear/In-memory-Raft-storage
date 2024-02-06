package inmemory

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

// Создание хранилища
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
	}
}

// Get возвращает значение по ключу
func (s *InMemoryStore) Get(key string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, ok := s.data[key]
	return value, ok
}

// handleGet обрабатывает HTTP GET запросы
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

// Put добавляет или обновляет значение по ключу
func (s *InMemoryStore) Put(key, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = value
}

// handlePut обрабатывает HTTP PUT запросы
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

func main() {
	store := NewInMemoryStore()

	http.HandleFunc("/get", handleGet(store))
	http.HandleFunc("/put", handlePut(store))

	log.Printf("Server is running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
