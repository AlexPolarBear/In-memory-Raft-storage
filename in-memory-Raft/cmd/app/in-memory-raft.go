package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"inmemoryraft/internal/services"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type commandKind uint8

const (
	putCommand commandKind = iota
	getCommand
	deleteCommand
)

type raftNode struct {
	store *services.InMemoryStore
}

func (r *raftNode) Restore(snapshot io.ReadCloser) error {
	panic("unimplemented")
}

func (r *raftNode) Snapshot() (raft.FSMSnapshot, error) {
	panic("unimplemented")
}

type Command struct {
	Type  commandKind
	Key   string
	Value string
}

func (r *raftNode) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		fmt.Printf("Failed to unmarshal command: %s", err)
		return nil
	}

	result := make(chan string)

	switch cmd.Type {
	case putCommand:
		r.store.Put(cmd.Key, cmd.Value)
	case getCommand:
		r.store.Get(cmd.Key, result)
		value := <-result
		return []byte(value)
	case deleteCommand:
		r.store.Delete(cmd.Key)
	default:
		fmt.Printf("Unknown command type: %d", cmd.Type)
	}

	return nil
}

type RaftLogStore struct {
	logs map[uint64]*raft.Log
}

func NewRaftLogStore() *RaftLogStore {
	return &RaftLogStore{
		logs: make(map[uint64]*raft.Log),
	}
}

func (r *RaftLogStore) StoreLog(log *raft.Log) error {
	r.logs[log.Index] = log
	return nil
}

func (r *RaftLogStore) GetLog(index uint64, _ *raft.Log) (*raft.Log, error) {
	return r.logs[index], nil
}

func (r *RaftLogStore) StoreLogs(logs []*raft.Log) error {
	for _, log := range logs {
		r.logs[log.Index] = log
	}
	return nil
}

func (r *RaftLogStore) FirstIndex() (uint64, error) {
	if len(r.logs) == 0 {
		return 0, nil
	}
	var min uint64 = math.MaxUint64
	for index := range r.logs {
		if index < min {
			min = index
		}
	}
	return min, nil
}

func (r *RaftLogStore) LastIndex() (uint64, error) {
	var max uint64 = 0
	for index := range r.logs {
		if index > max {
			max = index
		}
	}
	return max, nil
}

func (r *RaftLogStore) GetLogEntries(index uint64, maxEntries uint64) ([]*raft.Log, error) {
	var entries []*raft.Log

	lastIndex, err := r.LastIndex()
	if err != nil {
		return nil, err
	}

	for i := index; i <= lastIndex && uint64(len(entries)) < maxEntries; i++ {
		if log, ok := r.logs[i]; ok {
			entries = append(entries, log)
		}
	}

	return entries, nil
}

func (r *RaftLogStore) DeleteRange(min, max uint64) error {
	for i := min; i <= max; i++ {
		delete(r.logs, i)
	}
	return nil
}

type RaftStableStore struct {
	Data map[string][]byte
}

func NewRaftStableStore() *RaftStableStore {
	return &RaftStableStore{
		Data: make(map[string][]byte),
	}
}

func (r *RaftStableStore) Put(key []byte, val []byte) error {
	r.Data[string(key)] = val
	return nil
}

func (r *RaftStableStore) Get(key []byte) ([]byte, error) {
	return r.Data[string(key)], nil
}

func (r *RaftStableStore) SetUint64(key []byte, val uint64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, val)
	return r.Put(key, buf)
}

func (r *RaftStableStore) GetUint64(key []byte) (uint64, error) {
	val, err := r.Get(key)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(val), nil
}

type PathsConfig struct {
	DataFile  string `json:"data_file"`
	IndexFile string `json:"index_file"`
}

func loadDataFromFile(s *services.InMemoryStore, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &s.Data)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func main() {
	store := services.NewInMemoryStore()
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "raft",
		Output: os.Stderr,
		Level:  hclog.LevelFromString("DEBUG"),
	})

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID("node1")
	raftConfig.Logger = logger

	logStore, err := raftboltdb.NewBoltStore("raft_log.log")
	if err != nil {
		log.Fatal(err)
	}
	defer logStore.Close()

	stableStore, err := raftboltdb.NewBoltStore("raft_stable.log")
	if err != nil {
		log.Fatal(err)
	}
	defer stableStore.Close()

	snapshotStore := raft.NewDiscardSnapshotStore()

	fsm := &raftNode{store: store}

	transport, err := raft.NewTCPTransport("localhost:7000", nil, 3, 10*time.Second, os.Stderr)
	if err != nil {
		log.Fatal(err)
	}

	raftNode, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			appliedIndex := raftNode.AppliedIndex()
			for i := raftNode.LastIndex(); i <= appliedIndex; i++ {
				logEntry, err := NewRaftLogStore().GetLogEntries(i, 10)
				if err != nil {
					log.Printf("Failed to get applied log at index %d: %s", i, err)
					continue
				}
				if logEntry == nil || len(logEntry) == 0 {
					continue
				}
				command := logEntry[len(logEntry)-1]
				// if !ok {
				// 	log.Printf("Failed to assert command type")
				// 	continue
				// }
				log.Printf("Applied command: %+v", command)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raftConfig.LocalID,
				Address: transport.LocalAddr(),
			},
		},
	}
	raftNode.BootstrapCluster(configuration)

	file, err := os.Open("in-memory/configs/config.json")
	// file, err := os.Open("../configs/config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := PathsConfig{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(config.DataFile); errors.Is(err, os.ErrNotExist) {
		_, err := os.Create(config.DataFile)
		if err != nil {
			log.Fatalf("Failed to create json file: %v", err)
		}
	}

	err = loadDataFromFile(store, config.DataFile)
	if err != nil {
		log.Fatalf("Failed to load data from file: %v", err)
	}

	// interval := 30 * time.Minute
	// go services.PeriodicSave(store, interval)
	// go services.Snapshot(store, interval)

	http.HandleFunc("/get", services.HandleGet(store))
	http.HandleFunc("/put", services.HandlePut(store))
	http.HandleFunc("/delete", services.HandleDelete(store))

	http.Handle("/", http.FileServer(http.Dir(config.IndexFile)))

	log.Printf("Server is running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
