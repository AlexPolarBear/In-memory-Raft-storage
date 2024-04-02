package services

import (
	"encoding/json"
	"fmt"
	"os"
)

func (s *InMemoryStore) PersistLogToFile() error {
	logFile, err := os.OpenFile("in-memory/internal/data/log.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла журнала: %w", err)
	}
	defer logFile.Close()

	logData, err := json.MarshalIndent(s.OperationLog, "", "    ")
	if err != nil {
		return fmt.Errorf("ошибка маршализации журнала операций: %w", err)
	}

	_, err = logFile.Write(logData)
	if err != nil {
		return fmt.Errorf("ошибка записи журнала в файл: %w", err)
	}

	return nil
}

// s.OperationLog = append(s.OperationLog, fmt.Sprintf("Set(%s, %s)", key, value))
// 	log.Printf("Установлено значение для ключа %s: %s", key, value)
