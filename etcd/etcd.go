package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// KeyValueStore представляет key-value хранилище
type KeyValueStore struct {
	data map[string]string
	mu   sync.RWMutex
}

// NewKeyValueStore создает новое key-value хранилище
func NewKeyValueStore() *KeyValueStore {
	return &KeyValueStore{
		data: make(map[string]string),
	}
}

// Get возвращает значение по ключу из хранилища
func (kv *KeyValueStore) Get(key string) (string, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	value, ok := kv.data[key]
	return value, ok
}

// Put добавляет или обновляет значение в хранилище по ключу
func (kv *KeyValueStore) Put(key, value string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[key] = value
}

// handleGet обрабатывает HTTP GET запросы
func handleGet(kv *KeyValueStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		value, ok := kv.Get(key)
		if !ok {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "Value for key '%s': %s", key, value)
	}
}

// handlePut обрабатывает HTTP PUT запросы
func handlePut(kv *KeyValueStore) http.HandlerFunc {
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

		kv.Put(requestData.Key, requestData.Value)
		fmt.Fprintf(w, "Successfully stored value for key '%s'", requestData.Key)
	}
}

func main() {
	kv := NewKeyValueStore()

	http.HandleFunc("/get", handleGet(kv))
	http.HandleFunc("/put", handlePut(kv))

	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}
