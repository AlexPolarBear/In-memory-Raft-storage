package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.etcd.io/etcd/clientv3"
)

const (
	baseURL = "http://localhost:8000"
	key     = "test_key"
	value   = "test_value"
)

func TestSetAndGet(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
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

	if len(resp.Kvs) != 1 {
		t.Errorf("Expected 1 key-value pair, got %d", len(resp.Kvs))
	}
	if string(resp.Kvs[0].Value) != value {
		t.Errorf("Expected value %s, got %s", value, resp.Kvs[0].Value)
	}
}

func BenchmarkPut(b *testing.B) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
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
		Endpoints: []string{"localhost:2379"},
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
		Endpoints: []string{"localhost:2379"},
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

func BenchmarkAPIPut(b *testing.B) {
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

func BenchmarkAPIGet(b *testing.B) {
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

func BenchmarkAPIDelete(b *testing.B) {
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

// func main() {
// 	fmt.Println("Running benchmarks...")
// 	client, err := clientv3.New(clientv3.Config{
// 		Endpoints: []string{"localhost:2379"},
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	defer client.Close()

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	if _, err := client.Delete(ctx, key); err != nil {
// 		fmt.Println("Error deleting key from etcd:", err)
// 	}

// 	testing.Benchmark(BenchmarkDelete)
// 	testing.Benchmark(BenchmarkPut)
// 	testing.Benchmark(BenchmarkGet)

// 	testing.Benchmark(BenchmarkAPIPut)
// 	testing.Benchmark(BenchmarkAPIGet)
// 	testing.Benchmark(BenchmarkAPIDelete)
// }

// type KeyValue struct {
// 	Key   string `json:"key"`
// 	Value string `json:"value"`
// }
