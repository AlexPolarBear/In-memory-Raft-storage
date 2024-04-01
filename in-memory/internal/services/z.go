package services

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"sort"
// 	"sync"
// 	"time"
// )

// func Snapshot(s *InMemoryStore, interval time.Duration) {
// 	ticker := time.NewTicker(interval)
// 	defer ticker.Stop()

// 	defer s.Mutex.RUnlock()

// 	snapshots := make([]map[string]string, 0)
// 	snapshotsMu := sync.Mutex{}

// 	for range ticker.C {
// 		s.Mutex.RLock()

// 		snapshot := make(map[string]string)
// 		for k, v := range s.Data {
// 			snapshot[k] = v
// 		}

// 		snapshotsMu.Lock()
// 		defer snapshotsMu.Unlock()

// 		snapshots = append(snapshots, snapshot)
// 		if len(snapshots) > 10 {
// 			snapshots = snapshots[1:]
// 		}

// 		// сохраняем снимок в файл
// 		timestamp := time.Now().Format("2006-01-02-15-04-05")
// 		filename := fmt.Sprintf("snapshot-%s.json", timestamp)
// 		SaveSnapshotToFile(s, filepath.Join("in-memory/internal/data/snapshots", filename), snapshot)
// 	}
// }

// func SaveSnapshotToFile(s *InMemoryStore, filename string, snapshot map[string]string) {
// 	file, err := os.Create(filename)
// 	if err != nil {
// 		log.Println("Error to create file:", err)
// 		return
// 	}
// 	defer file.Close()

// 	enc := json.NewEncoder(file)
// 	if err := enc.Encode(snapshot); err != nil {
// 		log.Println("Error to write in file:", err)
// 		return
// 	}
// 	log.Println("Snapshot successfully saved in file:", filename)
// }

// // GetSnapshots возвращает последние 10 снимков, отсортированные по времени
// func GetSnapshots(dir string) ([]map[string]string, error) {
// 	files, err := ioutil.ReadDir(dir)
// 	if err != nil {
// 		return nil, err
// 	}

// 	snapshots := make([]map[string]string, 0)
// 	for _, file := range files {
// 		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
// 			snapshot, err := LoadSnapshotFromFile(filepath.Join(dir, file.Name()))
// 			if err != nil {
// 				return nil, err
// 			}
// 			snapshots = append(snapshots, snapshot)
// 		}
// 	}

// 	// сортируем снимки по времени
// 	sort.Slice(snapshots, func(i, j int) bool {
// 		return snapshots[i]["timestamp"] < snapshots[j]["timestamp"]
// 	})

// 	// возвращаем последние 10 снимков
// 	if len(snapshots) > 10 {
// 		snapshots = snapshots[len(snapshots)-10:]
// 	}
// 	return snapshots, nil
// }

// func LoadSnapshotFromFile(filename string) (map[string]string, error) {
// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	dec := json.NewDecoder(file)
// 	snapshot := make(map[string]string)
// 	if err := dec.Decode(&snapshot); err != nil {
// 		return nil, err
// 	}
// 	return snapshot, nil
// }
