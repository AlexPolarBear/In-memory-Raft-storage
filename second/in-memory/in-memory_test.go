package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()

	t.Run("Put", func(t *testing.T) {
		store.Put("test_key", "test_value")

		value, ok := store.Get("test_key")
		if !ok || value != "test_value" {
			t.Errorf("Expected value 'test_value' for key 'test_key', got '%s'", value)
		}
	})

	t.Run("Get", func(t *testing.T) {
		store.Put("test_key", "test_value")

		value, ok := store.Get("test_key")
		if !ok || value != "test_value" {
			t.Errorf("Expected value 'test_value' for key 'test_key', got '%s'", value)
		}

		_, ok = store.Get("nonexistent_key")
		if ok {
			t.Errorf("Expected key 'nonexistent_key' to be not found")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		store.Put("test_key", "test_value")
		store.Delete("test_key")

		_, ok := store.Get("test_key")
		if ok {
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
		store.Get(key)
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

	var wg sync.WaitGroup
	wg.Add(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := i
		go func() {
			defer wg.Done()
			url := fmt.Sprintf("%s/put", server.URL)
			data := fmt.Sprintf(`{"key": "key%d", "value": "value%d"}`, n, n)
			body := strings.NewReader(data)
			http.Post(url, "application/json", body) //nil
		}()
	}

	wg.Wait()
}

func BenchmarkHTTPGet(b *testing.B) {
	store := NewInMemoryStore()
	server := httptest.NewServer(http.HandlerFunc(handleGet(store)))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := i
		go func() {
			defer wg.Done()
			url := fmt.Sprintf("%s/get?key=key%d", server.URL, n)
			http.Get(url)
		}()
	}

	wg.Wait()
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

	var wg sync.WaitGroup
	wg.Add(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var k = i
		go func() {
			defer wg.Done()
			url := fmt.Sprintf("%s/delete?key=key%d", server.URL, k)
			req, _ := http.NewRequest(http.MethodDelete, url, nil)
			http.DefaultClient.Do(req)
		}()
	}

	wg.Wait()
}
