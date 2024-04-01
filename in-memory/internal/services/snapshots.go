package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func Snapshot(s *InMemoryStore, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	snapshots := make([]map[string]string, 0)
	snapshotsMutex := sync.Mutex{}

	for range ticker.C {
		SnapshotCreat(s, &snapshotsMutex, &snapshots)
	}
}

func SnapshotCreat(s *InMemoryStore, snapshotsMutex *sync.Mutex, snapshots *[]map[string]string) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	defer snapshotsMutex.Unlock()

	snapshot := make(map[string]string)
	for k, v := range s.Data {
		snapshot[k] = v
	}

	snapshotsMutex.Lock()
	*snapshots = append(*snapshots, snapshot)
	if len(*snapshots) > 10 {
		*snapshots = (*snapshots)[1:]
	}
	// snapshotsMutex.Unlock()

	go func() {
		s.SnapCh <- snapshot
	}()

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("snapshot-%s.json", timestamp)
	SaveSnapshotToFile(s, filepath.Join("in-memory/internal/data/snapshots", filename), snapshot)
	DeleteOldSnapshots(filepath.Join("in-memory/internal/data/snapshots"), 10)
}

func SaveSnapshotToFile(s *InMemoryStore, filename string, snapshot map[string]string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Println("Error to create file:", err)
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snapshot); err != nil {
		log.Println("Error to write in file:", err)
		return
	}
	log.Println("Snapshot successfully saved in file:", filename)
}

func GetSnapshots(dir string) ([]map[string]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	snapshots := make([]map[string]string, 0)
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			snapshot, err := LoadSnapshotFromFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			snapshots = append(snapshots, snapshot)
		}
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i]["timestamp"] < snapshots[j]["timestamp"]
	})

	if len(snapshots) > 10 {
		snapshots = snapshots[len(snapshots)-10:]
	}
	return snapshots, nil
}

func LoadSnapshotFromFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	snapshot := make(map[string]string)
	if err := dec.Decode(&snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func DeleteOldSnapshots(dir string, maxSnapshots int) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Println("Error reading directory:", err)
		return
	}

	sort.Slice(files, func(i, j int) bool {
		iTime, err := getTimeFromFilename(files[i].Name())
		if err != nil {
			return false
		}
		jTime, err := getTimeFromFilename(files[j].Name())
		if err != nil {
			return false
		}
		return iTime.Before(jTime)
	})

	for i := 0; i < len(files)-maxSnapshots; i++ {
		err := os.Remove(filepath.Join(dir, files[i].Name()))
		if err != nil {
			log.Println("Error deleting file:", err)
		}
	}
}

func getTimeFromFilename(filename string) (time.Time, error) {
	timestamp := strings.TrimSuffix(filename, ".json")
	timestamp = strings.ReplaceAll(timestamp, "snapshot-", "")
	t, err := time.Parse("2006-01-02_15-04-05", timestamp)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
