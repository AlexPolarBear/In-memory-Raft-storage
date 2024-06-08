package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func PostTestingFunc() *http.Response {
	values := map[string]string{"testKey": "testValue"}
	jsonData, _ := json.Marshal(values)

	resp, err := http.Post("http://localhost:8080/keys", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// _ = fmt.Errorf("failed to add test data, got %v", err)
		return resp
	}
	defer resp.Body.Close()
	return resp
}

func TestHandlePostKey(t *testing.T) {
	resp := PostTestingFunc()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestHandleGetKey(t *testing.T) {
	_ = PostTestingFunc()

	resp, err := http.Get("http://localhost:8081/keys/testKey")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data["testKey"] != "testValue" {
		t.Fatalf("Expected value 'testValue', got '%s'", data["testKey"])
	}
}

func TestHandleDeleteKey(t *testing.T) {
	_ = PostTestingFunc()

	req, _ := http.NewRequest("DELETE", "http://localhost:8080/keys/testKey", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestHandleJoin(t *testing.T) {
	values := map[string]string{"addr": "localhost:8081", "id": "node1"}
	jsonData, _ := json.Marshal(values)

	resp, err := http.Post("http://localhost:8080/join", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func BenchmarkHandlePostKey(b *testing.B) {
	values := map[string]string{"testKey": "testValue"}
	jsonData, _ := json.Marshal(values)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := http.Post("http://localhost:8080/keys", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			b.Fatalf("Error posting key: %v", err)
		}
	}
}

func BenchmarkHandleGetKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := http.Get("http://localhost:8080/keys/testKey")
		if err != nil {
			b.Fatalf("Error fetching key: %v", err)
		}
	}
}

func BenchmarkHandleDeleteKey(b *testing.B) {
	req, _ := http.NewRequest("DELETE", "http://localhost:8080/keys/testKey", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatalf("Error deleting key: %v", err)
		}
	}
}

func BenchmarkHandleJoin(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		port := 5000 + i
		if port == 8080 {
			continue
		}
		addr := fmt.Sprintf("localhost:%d", port)
		values := map[string]string{"addr": addr, "id": fmt.Sprintf("node%d", i)}
		jsonData, _ := json.Marshal(values)

		_, err := http.Post("http://localhost:8080/join", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			b.Fatalf("Error joining node: %v", err)
		}
	}
}
