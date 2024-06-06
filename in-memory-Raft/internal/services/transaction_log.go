package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Command   command   `json:"command"`
	ApplyTime time.Time `json:"applyTime"`
}

type TransactionLog struct {
	Entries []LogEntry `json:"entries"`
	mutex   sync.RWMutex
}

func NewTransactionLog() *TransactionLog {
	return &TransactionLog{}
}

func (tl *TransactionLog) Append(entry LogEntry) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	tl.Entries = append(tl.Entries, entry)
}

func (tl *TransactionLog) Save(filename string) error {
	tl.mutex.RLock()
	defer tl.mutex.RUnlock()
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(tl.Entries)
}

func (tl *TransactionLog) Load(filename string, store *InMemoryStore) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var entries []LogEntry
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&entries); err != nil {
		return err
	}

	for _, entry := range entries {
		switch entry.Command.Op {
		case "Put":
			if err := store.Put(entry.Command.Key, entry.Command.Value); err != nil {
				return err
			}
		case "Delete":
			if err := store.Delete(entry.Command.Key); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown command operation: %s", entry.Command.Op)
		}
	}

	return nil
}
