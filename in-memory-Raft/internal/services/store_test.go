package services

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestStoreOpen(t *testing.T) {
	store := NewStore()
	mkDir, _ := os.MkdirTemp("", "store_test")
	defer os.RemoveAll(mkDir)

	if store == nil {
		t.Fatal("failed to create store")
	}
	store.RaftBind = "localhost:8888"
	store.RaftDir = mkDir

	err := store.InitNode(false, "nodeTest")
	if err != nil {
		t.Fatalf("failed to init node: %s", err)
	}
}

func initTestNode() (*InMemoryStore, error) {
	store := NewStore()
	mkDir, _ := os.MkdirTemp("", "store_test")
	defer os.RemoveAll(mkDir)

	if store == nil {
		err := fmt.Errorf("failed to create store")
		return store, err
	}
	store.RaftBind = "localhost:8888"
	store.RaftDir = mkDir

	err := store.InitNode(true, "nodeTest")
	if err != nil {
		return store, err
	}

	// enshure that leader is chosen
	time.Sleep(3 * time.Second)
	return store, nil
}

// tests that a command can be applied to the log stored in RAM.
func TestStoreOperations(t *testing.T) {
	store, err := initTestNode()
	if err != nil {
		t.Fatalf("InitNode failed: %s", err)
	}

	tkey := "testkey"
	tvalue := "testvalue"

	err = store.Put(tkey, tvalue)
	if err != nil {
		t.Fatalf("failed to set key: %s", err.Error())
	}

	// Wait for committed log entry to be applied
	time.Sleep(500 * time.Millisecond)

	val, err := store.Get(tkey)
	if err != nil {
		t.Fatalf("failed to get key: %s", err.Error())
	}
	if val != tvalue {
		t.Fatalf("key has wrong value: %s", val)
	}

	err = store.Delete(tkey)
	if err != nil {
		t.Fatalf("failed to delete key: %s", err.Error())
	}

	// Wait for committed log entry to be applied
	time.Sleep(500 * time.Millisecond)
	val, _ = store.Get(tkey)
	if val != "" {
		t.Fatalf("key has wrong value: %s", val)
	}
}

func BenchmarkInMemoryStore_Put(b *testing.B) {
	store, err := initTestNode()
	if err != nil {
		b.Fatalf("InitNode failed: %s", err)
	}

	key := "testKey"
	value := "testValue"

	for i := 0; i < b.N; i++ {
		store.Put(key, value)
	}
}

func BenchmarkInMemoryStore_Get(b *testing.B) {
	store, err := initTestNode()
	if err != nil {
		b.Fatalf("InitNode failed: %s", err)
	}

	key := "testKey"
	value := "testValue"
	store.data[key] = value

	for i := 0; i < b.N; i++ {
		store.Get(key)
	}
}

func BenchmarkInMemoryStore_Delete(b *testing.B) {
	store, err := initTestNode()
	if err != nil {
		b.Errorf("InitNode failed: %s", err)
	}

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("testKey%d", i)
		value := fmt.Sprintf("testValue%d", i)
		store.data[key] = value
		store.Delete(key)
	}
}
