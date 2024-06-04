package services

import (
	"encoding/json"
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
	return encoder.Encode(tl.Entries)
}

func (tl *TransactionLog) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	return decoder.Decode(&tl.Entries)
}
