package inmemory

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

	// Тест для операции Put
	t.Run("Put", func(t *testing.T) {
		store.Put("test_key", "test_value")

		// Проверяем, что значение успешно добавлено в хранилище
		value, ok := store.Get("test_key")
		if !ok || value != "test_value" {
			t.Errorf("Expected value 'test_value' for key 'test_key', got '%s'", value)
		}
	})

	// Тест для операции Get
	t.Run("Get", func(t *testing.T) {
		store.Put("test_key", "test_value")

		// Проверяем, что получаем значение из хранилища
		value, ok := store.Get("test_key")
		if !ok || value != "test_value" {
			t.Errorf("Expected value 'test_value' for key 'test_key', got '%s'", value)
		}

		// Проверяем случай, когда ключ отсутствует
		_, ok = store.Get("nonexistent_key")
		if ok {
			t.Errorf("Expected key 'nonexistent_key' to be not found")
		}
	})
}

// Бенчмарк для операции Put
func BenchmarkPut(b *testing.B) {
	store := NewInMemoryStore()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		store.Put(key, value)
	}
}

// Бенчмарк для операции Get
func BenchmarkGet(b *testing.B) {
	store := NewInMemoryStore()

	// Заполняем хранилище данными
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		store.Put(key, value)
	}

	b.ResetTimer()

	// Выполняем операцию Get
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		store.Get(key)
	}
}

// Бенчмарк для HTTP GET запроса
func BenchmarkHTTPGet(b *testing.B) {
	store := NewInMemoryStore()
	server := httptest.NewServer(http.HandlerFunc(handleGet(store)))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(b.N)

	b.ResetTimer()

	// Выполняем параллельные HTTP GET запросы
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

// Бенчмарк для HTTP PUT запроса
func BenchmarkHTTPPut(b *testing.B) {
	store := NewInMemoryStore()
	server := httptest.NewServer(http.HandlerFunc(handlePut(store)))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(b.N)

	b.ResetTimer()

	// Выполняем параллельные HTTP PUT запросы
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
