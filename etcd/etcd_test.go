package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// Бенчмарк для операции Put в key-value хранилище
func BenchmarkPut(b *testing.B) {
	kv := NewKeyValueStore()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		kv.Put(key, value)
	}
}

// Бенчмарк для операции Get в key-value хранилище
func BenchmarkGet(b *testing.B) {
	kv := NewKeyValueStore()

	// Заполняем хранилище данными
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		kv.Put(key, value)
	}

	b.ResetTimer()

	// Выполняем операцию Get
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		kv.Get(key)
	}
}

// Бенчмарк для HTTP PUT запроса
func BenchmarkHTTPPut(b *testing.B) {
	kv := NewKeyValueStore()
	server := httptest.NewServer(http.HandlerFunc(handlePut(kv)))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(b.N)

	b.ResetTimer()

	// Выполняем параллельные HTTP PUT запросы
	for i := 0; i < b.N; i++ {
		go func() {
			defer wg.Done()
			url := fmt.Sprintf("%s/put", server.URL)
			// fmt.Sprintf(`{"key": "key%d", "value": "value%d"}`, i, i) //data
			http.Post(url, "application/json", nil)
		}()
	}

	wg.Wait()
}

// Бенчмарк для HTTP GET запроса
func BenchmarkHTTPGet(b *testing.B) {
	kv := NewKeyValueStore()
	server := httptest.NewServer(http.HandlerFunc(handleGet(kv)))
	defer server.Close()

	// Заполняем хранилище данными
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		kv.Put(key, value)
	}

	b.ResetTimer()

	// Выполняем параллельные HTTP GET запросы
	for i := 0; i < b.N; i++ {
		go func() {
			url := fmt.Sprintf("%s/get?key=key%d", server.URL, i)
			http.Get(url)
		}()
	}
}
