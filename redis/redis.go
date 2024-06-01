package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var rdb *redis.Client

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/data/{key}", GetDataHandler).Methods("GET")
	router.HandleFunc("/api/v1/data/{key}", PutDataHandler).Methods("PUT")
	router.HandleFunc("/api/v1/data/{key}", DeleteDataHandler).Methods("DELETE")

	log.Println("Server started on: http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func GetDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	val, err := rdb.Get(context.Background(), key).Result()
	if err == redis.Nil {
		http.Error(w, "key does not exist", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{key: val})
}

func PutDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = rdb.Set(context.Background(), key, data[key], 10*time.Minute).Err()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func DeleteDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := rdb.Del(context.Background(), key).Err()
	if err == redis.Nil {
		http.Error(w, "key does not exist", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
