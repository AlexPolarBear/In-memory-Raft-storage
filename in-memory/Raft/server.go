package raft

// import "github.com/hashicorp/raft"

// func NewRaftServer(config raft.Config, store *InMemoryRaft) (*raft.Raft, error) {
// 	// Создайте транспорт для взаимодействия с другими узлами
// 	server, transport := raft.NewInmemTransport()

// 	// Создайте сервер Raft
// 	raftServer, err := raft.NewRaft(config, store, transport)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Запустите сервер Raft
// 	go raftServer.Start()

// 	return raftServer, nil
// }
