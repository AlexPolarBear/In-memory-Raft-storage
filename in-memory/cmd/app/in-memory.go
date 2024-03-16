package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"inmemory/internal/services"
)

type PathsConfig struct {
	DataFile  string `json:"data_file"`
	IndexFile string `json:"index_file"`
	// LogFile   string `json:"log_file`
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

	file, err := os.Open("in-memory/configs/config.json")
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

	// if _, err := os.Stat(config.LogFile); errors.Is(err, os.ErrNotExist) {
	// 	_, err := os.Create(config.LogFile)
	// 	if err != nil {
	// 		log.Fatalf("Failed to create log file: %v", err)
	// 	}
	// } else {
	// 	err = loadLogFromFile(store, config.LogFile)
	// 	if err != nil {
	// 		log.Fatalf("Failed to load log from file: %v", err)
	// 	}
	// }

	err = loadDataFromFile(store, config.DataFile)
	if err != nil {
		log.Fatalf("Failed to load data from file: %v", err)
	}

	interval := 24 * time.Hour
	go services.PeriodicSave(store, interval)

	http.HandleFunc("/get", services.HandleGet(store))
	http.HandleFunc("/put", services.HandlePut(store))
	http.HandleFunc("/delete", services.HandleDelete(store))

	http.Handle("/", http.FileServer(http.Dir(config.IndexFile)))

	log.Printf("Server is running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
