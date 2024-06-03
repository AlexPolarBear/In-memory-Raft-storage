package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"inmemoryraft/internal/api"
	"inmemoryraft/internal/services"
	"log"
	"net/http"
	"os"
	"os/signal"
)

const (
	DefaultHTTPAddr = "localhost:8000"
	DefaultRaftAddr = "localhost:7000"
)

var httpAddr string
var raftAddr string
var joinAddr string
var nodeID string

func init() {
	flag.StringVar(&httpAddr, "haddr", DefaultHTTPAddr, "Set the HTTP bind address")
	flag.StringVar(&raftAddr, "raddr", DefaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&joinAddr, "join", "", "Set join address, if any")
	flag.StringVar(&nodeID, "id", "", "Node ID. If not set, same as Raft bind address")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <raft-data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}

	if nodeID == "" {
		nodeID = raftAddr
	}

	raftDir := "internal/data/snapshots/" + flag.Arg(0)
	if raftDir == "" {
		log.Fatalln("No Raft storage directory specified")
	}
	if err := os.MkdirAll(raftDir, 0700); err != nil {
		log.Fatalf("failed to create path for Raft storage: %s", err.Error())
	}

	store := services.New()
	store.RaftDir = raftDir
	store.RaftBind = raftAddr
	if err := store.InitNode(joinAddr == "", nodeID); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	h := api.New(httpAddr, store)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	if joinAddr != "" {
		if err := join(joinAddr, raftAddr, nodeID); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	log.Printf("raft node started successfully, listening on http://%s", httpAddr)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("raft node exiting")
}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
