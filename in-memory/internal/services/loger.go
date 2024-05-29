package services

import (
	"encoding/json"
	"fmt"
	"os"
)

func (s *InMemoryStore) PersistLogToFile() error {
	logFile, err := os.OpenFile("in-memory/internal/data/log.log",
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening the log file occurred: %w", err)
	}
	defer logFile.Close()

	logData, err := json.MarshalIndent(s.OperationLog, "", "    ")
	if err != nil {
		return fmt.Errorf("operation log marshallization error occurred: %w", err)
	}

	_, err = logFile.Write(logData)
	if err != nil {
		return fmt.Errorf("error writing the log to the file occurred: %w", err)
	}

	return nil
}
