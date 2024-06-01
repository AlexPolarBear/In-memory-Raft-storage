package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var (
	router *mux.Router
)

func setup() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	router = mux.NewRouter()
	router.HandleFunc("/api/v1/data/{key}", GetDataHandler).Methods("GET")
	router.HandleFunc("/api/v1/data/{key}", PutDataHandler).Methods("PUT")
	router.HandleFunc("/api/v1/data/{key}", DeleteDataHandler).Methods("DELETE")
}

func TestGetDataHandler(t *testing.T) {
	setup()

	ts := httptest.NewServer(router)
	defer ts.Close()

	key := "testKey"
	value := "testValue"
	rdb.Set(context.Background(), key, value, 0)

	resp, err := http.Get(ts.URL + "/api/v1/data/" + key)
	if err != nil {
		t.Fatalf("could not send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.StatusCode)
	}

	var data map[string]string
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if data[key] != value {
		t.Errorf("expected %s; got %s", value, data[key])
	}
}

func TestPutDataHandler(t *testing.T) {
	setup()

	ts := httptest.NewServer(router)
	defer ts.Close()

	data := map[string]string{"testKey": "newTestValue"}
	jsonData, _ := json.Marshal(data)

	req, _ := http.NewRequest("PUT", ts.URL+"/api/v1/data/testKey", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("could not send PUT request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status Created; got %v", resp.StatusCode)
	}

	val, err := rdb.Get(context.Background(), "testKey").Result()
	if err != nil {
		t.Fatalf("could not get data from Redis: %v", err)
	}
	if val != "newTestValue" {
		t.Errorf("expected %s; got %s", "newTestValue", val)
	}
}

func TestDeleteDataHandler(t *testing.T) {
	setup()

	ts := httptest.NewServer(router)
	defer ts.Close()

	key := "testKey"
	value := "testValue"
	rdb.Set(context.Background(), key, value, 0)

	req, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/data/"+key, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("could not send DELETE request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.StatusCode)
	}

	val, err := rdb.Get(context.Background(), key).Result()
	if err != redis.Nil {
		t.Errorf("expected key to be deleted; got %v", val)
	}
}

func BenchmarkGetDataHandler(b *testing.B) {
	setup()

	ts := httptest.NewServer(router)
	defer ts.Close()

	key := "testKey"
	value := "testValue"
	rdb.Set(context.Background(), key, value, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(ts.URL + "/api/v1/data/" + key)
		if err != nil {
			b.Fatalf("could not send GET request: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkPutDataHandler(b *testing.B) {
	setup()

	ts := httptest.NewServer(router)
	defer ts.Close()

	data := map[string]string{"testKey": "newTestValue"}
	jsonData, _ := json.Marshal(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("PUT", ts.URL+"/api/v1/data/testKey", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatalf("could not send PUT request: %v", err)
		}
		resp.Body.Close()
	}
}

func BenchmarkDeleteDataHandler(b *testing.B) {
	setup()

	ts := httptest.NewServer(router)
	defer ts.Close()

	key := "testKey"
	value := "testValue"
	rdb.Set(context.Background(), key, value, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("DELETE", ts.URL+"/api/v1/data/"+key, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatalf("could not send DELETE request: %v", err)
		}
		resp.Body.Close()
	}
}
