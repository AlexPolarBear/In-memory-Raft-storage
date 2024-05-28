package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/clientv3"
)

const (
	hostURL = "localhost:2379"
	baseURL = "http://localhost:8000"
	key     = "test_key"
	value   = "test_value"
)

func TestPutAndGet(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{hostURL},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	_, err = client.Put(context.Background(), key, value)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Get(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(resp.Kvs), "Expected 1 key-value pair")
	assert.Equal(t, value, string(resp.Kvs[0].Value), "Expected value to be 'test_value'")
}

func TestGet(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{hostURL},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	_, err = client.Put(context.Background(), key, value)
	if err != nil {
		t.Fatal(err)
	}

	getReq, err := http.NewRequest("GET", fmt.Sprintf("/get?key=%s", key), nil)
	if err != nil {
		t.Fatal(err)
	}

	getRR := httptest.NewRecorder()
	getHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		resp, err := client.Get(context.Background(), key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(resp.Kvs) == 0 {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}

		kv := KeyValue{
			Key:   string(resp.Kvs[0].Key),
			Value: string(resp.Kvs[0].Value),
		}

		json.NewEncoder(w).Encode(kv)
	})

	getHandler.ServeHTTP(getRR, getReq)

	assert.Equal(t, http.StatusOK, getRR.Code)
}

func TestPut(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{hostURL},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	putReq, err := http.NewRequest("PUT", "/put", strings.NewReader(
		fmt.Sprintf(`{"key":"%s","value":"%s"}`, key, value)))
	if err != nil {
		t.Fatal(err)
	}

	putRR := httptest.NewRecorder()
	putHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var kv KeyValue
		if err := json.NewDecoder(r.Body).Decode(&kv); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := client.Put(context.Background(), kv.Key, kv.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Key %s set to %s", kv.Key, kv.Value)
	})

	putHandler.ServeHTTP(putRR, putReq)

	assert.Equal(t, http.StatusOK, putRR.Code)
	assert.Equal(t, fmt.Sprintf("Key %s set to %s", key, value), putRR.Body.String())
}

func TestDelete(t *testing.T) {

	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{hostURL},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("/delete?key=%s", key), nil)
	if err != nil {
		t.Fatal(err)
	}

	deleteRR := httptest.NewRecorder()
	deleteHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
			return
		}

		_, err := client.Delete(context.Background(), key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Key %s deleted", key)
	})

	deleteHandler.ServeHTTP(deleteRR, deleteReq)

	assert.Equal(t, http.StatusOK, deleteRR.Code)
	assert.Equal(t, fmt.Sprintf("Key %s deleted", key), deleteRR.Body.String())
}

func BenchmarkPut(b *testing.B) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{hostURL},
	})
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Put(context.Background(), fmt.Sprintf("%s_%d", key, i), value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{hostURL},
	})
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	for i := 0; i < b.N; i++ {
		_, err := client.Get(context.Background(), key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDelete(b *testing.B) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{hostURL},
	})
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	key := fmt.Sprintf("%s_delete", key)
	value := "test_value"

	_, err = client.Put(context.Background(), key, value)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Delete(context.Background(), key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestHTTPPut(t *testing.T) {
	payload := fmt.Sprintf(`{"key": "%s", "value": "%s"}`, key, value)

	resp, err := http.Post(baseURL+"/put", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestHTTPGet(t *testing.T) {
	resp, err := http.Get(baseURL + "/get?key=" + key)
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	var kv KeyValue
	if err := json.Unmarshal(body, &kv); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPDelete(t *testing.T) {
	key := fmt.Sprintf("%s_delete_api", key)
	value := "test_value"

	payload := fmt.Sprintf(`{"key": "%s", "value": "%s"}`, key, value)
	resp, err := http.Post(baseURL+"/put", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	resp, err = http.Get(baseURL + "/delete?key=" + key)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func BenchmarkHTTPPut(b *testing.B) {
	payload := fmt.Sprintf(`{"key": "%s", "value": "%s"}`, key, value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(baseURL+"/put", "application/json", strings.NewReader(payload))
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkHTTPGet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(baseURL + "/get?key=" + key)
		if err != nil {
			b.Fatal(err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			b.Fatal(err)
		}

		var kv KeyValue
		if err := json.Unmarshal(body, &kv); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTPDelete(b *testing.B) {
	key := fmt.Sprintf("%s_delete_api", key)
	value := "test_value"

	payload := fmt.Sprintf(`{"key": "%s", "value": "%s"}`, key, value)
	resp, err := http.Post(baseURL+"/put", "application/json", strings.NewReader(payload))
	if err != nil {
		b.Fatal(err)
	}
	resp.Body.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(baseURL + "/delete?key=" + key)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
