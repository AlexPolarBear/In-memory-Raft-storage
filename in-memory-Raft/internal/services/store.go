package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

type command struct {
	Op    string `json:"op:omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type InMemoryStore struct {
	RaftDir  string
	RaftBind string // localhost:7000

	data  map[string]string
	mutex sync.RWMutex

	raft *raft.Raft

	logger *log.Logger
}

func New() *InMemoryStore {
	return &InMemoryStore{
		data:   make(map[string]string),
		logger: log.New(os.Stderr, "[store] ", log.LstdFlags),
	}
}

func (ims *InMemoryStore) InitNode(enableSingle bool, localID string) error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)

	addr, err := net.ResolveTCPAddr("tcp", ims.RaftBind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(ims.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	snapshots, err := raft.NewFileSnapshotStore(ims.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	ra, err := raft.NewRaft(config, (*fsm)(ims), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("error in creat node: %s", err)
	}
	ims.raft = ra

	if enableSingle {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return nil
}

func (ims *InMemoryStore) Get(key string) (string, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()
	value, ok := ims.data[key]
	if !ok {
		return "", fmt.Errorf("key '%s' not found", key)
	}
	return value, nil
}

func (ims *InMemoryStore) Put(key, value string) error {
	if ims.raft.State() != raft.Leader {
		return fmt.Errorf("not leader")
	}

	c := &command{
		Op:    "set",
		Key:   key,
		Value: value,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := ims.raft.Apply(b, raftTimeout)
	return f.Error()
}

func (ims *InMemoryStore) Delete(key string) error {
	if ims.raft.State() != raft.Leader {
		return fmt.Errorf("not leader")
	}

	c := &command{
		Op:  "delete",
		Key: key,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := ims.raft.Apply(b, raftTimeout)
	return f.Error()
}

func (ims *InMemoryStore) Join(nodeID, addr string) error {
	ims.logger.Printf("received join request for remote node %s at %s", nodeID, addr)

	configFuture := ims.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		ims.logger.Printf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				ims.logger.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := ims.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := ims.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	ims.logger.Printf("node %s at %s joined successfully", nodeID, addr)
	return nil
}

type fsm InMemoryStore

func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command

	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Op {
	case "set":
		return f.applySet(c.Key, c.Value)
	case "delete":
		return f.applyDelete(c.Key)
	default:
		panic(fmt.Sprintf("unrecognized command op: %s", c.Op))
	}
}

func (f *fsm) applySet(key, value string) interface{} {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.data[key] = value
	return nil
}

func (f *fsm) applyDelete(key string) interface{} {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	delete(f.data, key)
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	dataCopy := make(map[string]string)
	for k, v := range f.data {
		dataCopy[k] = v
	}
	return &fsmSnapshot{store: dataCopy}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	o := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&o); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	f.data = o
	return nil
}

type fsmSnapshot struct {
	store map[string]string
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		b, err := json.Marshal(f.store)
		if err != nil {
			return err
		}

		if _, err := sink.Write(b); err != nil {
			return err
		}

		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapshot) Release() {}
