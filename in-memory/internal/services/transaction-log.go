package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type LogEntry struct {
	Key       string
	Value     string
	Time      time.Time
	Operation string
}

func (s *InMemoryStore) LogOperation(op, key, value string) {
	if s == nil {
		log.Println("Store is nil")
		return
	}
	if s.TransactionLog == nil {
		log.Println("TransactionLog is nil")
		return
	}

	entry := LogEntry{
		Time:      time.Now(),
		Operation: op,
		Key:       key,
		Value:     value,
	}
	s.TransactionLog.Transactions = append(s.TransactionLog.Transactions, entry)
	if err := s.PersistLogToFile(); err != nil {
		log.Printf("Failed to persist log to file: %v", err)
	}
}

func (s *InMemoryStore) RestoreState() error {
	snapshots, err := GetSnapshots("in-memory/internal/data/snapshots")
	if err != nil {
		return err
	}

	if len(snapshots) > 0 {
		s.Data = snapshots[len(snapshots)-1]
	}

	operations, err := LoadOperations("in-memory/internal/data/log.log")
	if err != nil {
		return err
	}

	for _, op := range operations {
		switch op.Operation {
		case "PUT":
			s.Data[op.Key] = op.Value
		case "DELETE":
			delete(s.Data, op.Key)
		}
	}

	return nil
}

func LoadOperations(filename string) ([]LogEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening log file: %w", err)
	}
	defer file.Close()

	var entries []LogEntry
	dec := json.NewDecoder(file)
	if err := dec.Decode(&entries); err != nil {
		return nil, fmt.Errorf("error decoding log entries: %w", err)
	}

	return entries, nil
}

func Rollback(currentState map[string]string, operations []LogEntry) map[string]string {
	newState := make(map[string]string)
	for k, v := range currentState {
		newState[k] = v
	}

	for i := len(operations) - 1; i >= 0; i-- {
		op := operations[i]
		switch op.Operation {
		case "PUT":
			newState[op.Key] = op.Value
		case "DELETE":
			delete(newState, op.Key)
		}
	}

	return newState
}

func HandlerRollback(s *InMemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		operations, err := LoadOperations("in-memory/internal/data/log.log")
		if err != nil {
			http.Error(w, "Failed to load operations from log file", http.StatusInternalServerError)
			return
		}

		currentState := make(map[string]string)

		newState := Rollback(currentState, operations)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newState)

		r.Body.Close()
	}
}
