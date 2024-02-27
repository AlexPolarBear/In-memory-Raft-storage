package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()

	t.Run("Put", func(t *testing.T) {
		store.Put("test_key", "test_value")

		time.Sleep(100 * time.Millisecond) // Ждем завершения асинхронной операции

		result := make(chan string)
		go store.Get("test_key", result)
		value := <-result
		if value != "test_value" {
			t.Errorf("Expected value 'test_value' for key 'test_key', got '%s'", value)
		}
	})

	t.Run("Get", func(t *testing.T) {
		store.Put("test_key", "test_value")

		time.Sleep(100 * time.Millisecond) // Ждем завершения асинхронной операции

		result := make(chan string)
		go store.Get("test_key", result)
		value := <-result

		if value != "test_value" {
			t.Errorf("Expected value 'test_value' for key 'test_key', got '%s'", value)
		}

		time.Sleep(100 * time.Millisecond) // Ждем завершения асинхронной операции

		result = make(chan string)
		go store.Get("nonexistent_key", result)
		value = <-result

		if value != "" {
			t.Errorf("Expected key 'nonexistent_key' to be not found")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		store.Put("test_key", "test_value")

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			store.Delete("test_key")
		}()

		wg.Wait() // Ждем завершения операции Delete

		time.Sleep(100 * time.Millisecond) // Даем немного времени на завершение асинхронной операции

		result := make(chan string)
		go store.Get("test_key", result)
		value := <-result

		if value != "" {
			t.Errorf("Expected key 'test_key' to be deleted")
		}
	})
}

func BenchmarkPut(b *testing.B) {
	store := NewInMemoryStore()

	// f, err := os.Create("cpu_profile.csv")
	// if err != nil {
	// 	b.Fatalf("Failed to create CPU profile: %v", err)
	// }
	// defer f.Close()

	// err = pprof.StartCPUProfile(f)
	// if err != nil {
	// 	b.Fatalf("Failed to start CPU profile: %v", err)
	// }
	// defer pprof.StopCPUProfile()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		store.Put(key, value)
	}
}

func BenchmarkGet(b *testing.B) {
	store := NewInMemoryStore()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		store.Put(key, value)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		result := make(chan string)
		go store.Get(key, result)
		<-result
	}
}

func BenchmarkDelete(b *testing.B) {
	store := NewInMemoryStore()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		store.Put(key, value)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		store.Delete(key)
	}
}

func BenchmarkHTTPPut(b *testing.B) {
	store := NewInMemoryStore()
	server := httptest.NewServer(http.HandlerFunc(handlePut(store)))
	defer server.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := i
		url := fmt.Sprintf("%s/put", server.URL)
		data := fmt.Sprintf(`{"key": "key%d", "value": "value%d"}`, n, n)
		body := strings.NewReader(data)
		http.Post(url, "application/json", body)
	}
}

func BenchmarkHTTPGet(b *testing.B) {
	store := NewInMemoryStore()
	server := httptest.NewServer(http.HandlerFunc(handleGet(store)))
	defer server.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := i
		url := fmt.Sprintf("%s/get?key=key%d", server.URL, n)
		http.Get(url)
	}
}

func BenchmarkHTTPDelete(b *testing.B) {
	store := NewInMemoryStore()
	server := httptest.NewServer(http.HandlerFunc(handleDelete(store)))
	defer server.Close()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		store.Put(key, value)
	}

	b.ResetTimer()

	var wg sync.WaitGroup
	wg.Add(b.N)

	for i := 0; i < b.N; i++ {
		k := i
		go func() {
			defer wg.Done()
			url := fmt.Sprintf("%s/delete?key=key%d", server.URL, k)
			req, _ := http.NewRequest(http.MethodDelete, url, nil)
			http.DefaultClient.Do(req)
		}()
	}

	wg.Wait()
}
