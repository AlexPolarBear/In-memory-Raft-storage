package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.etcd.io/etcd/clientv3"
)

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var (
	etcdEndpoints = []string{"localhost:2379"} // список узлов etcd
)

func NewEtcd() {

}

func main() {
	// Создаем клиент etcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints: etcdEndpoints,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
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

	log.Printf("Server is running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
