package services

import (
	"encoding/json"

	"github.com/hashicorp/raft"
)

type Snapshot struct {
	Store *InMemoryStore
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	data, err := json.Marshal(s.Store.Data)
	if err != nil {
		return err
	}

	if _, err := sink.Write(data); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (s *Snapshot) Release() {
	panic("Implement me!")
}
