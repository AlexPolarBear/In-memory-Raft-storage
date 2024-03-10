package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"inmemory/internal/services"
)

func loadDataFromFile(store *services.InMemoryStore, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	_ = decoder.Decode(&store.Data)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return err
	// }

	return nil
}

func main() {
	store := services.NewInMemoryStore()

	filename := "data.json"

	if _, err := os.Stat(""); errors.Is(err, os.ErrNotExist) {
		_, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Failed to create json file: %v", err)
		}
	}

	err := loadDataFromFile(store, "data.json")
	if err != nil {
		log.Fatalf("Failed to load data from file: %v", err)
	}

	http.HandleFunc("/get", services.HandleGet(store))
	http.HandleFunc("/put", services.HandlePut(store))
	http.HandleFunc("/delete", services.HandleDelete(store))

	http.Handle("/", http.FileServer(http.Dir("../../configs")))

	log.Printf("Server is running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
