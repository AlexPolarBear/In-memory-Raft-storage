package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"inmemory/internal/services"
)

func HandleGet(s *services.InMemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		value, ok := s.Get(key)
		if !ok {
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

func HandlePut(s *services.InMemoryStore) http.HandlerFunc {
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

func HandleDelete(s *services.InMemoryStore) http.HandlerFunc {
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
