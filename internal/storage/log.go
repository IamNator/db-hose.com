package storage

import (
	"bytes"
	"fmt"
	"time"

	"dbhose/internal/domain"
	"dbhose/pkg"
	"encoding/json"
)

func (sm *StorageManager) LogBackup(dur time.Duration, email, affectedFile string) error {
	return sm.logOperation("Backup", dur, email, "creating backup", affectedFile)
}

func (sm *StorageManager) LogRestore(dur time.Duration, email, affectedFile string) error {
	return sm.logOperation("Restore", dur, email, "restoring backup", affectedFile)
}

func (sm *StorageManager) logOperation(operation string, dur time.Duration, email, description, affectedFile string) error {
	currentDate := time.Now().UTC().Format("2006_01_02")
	logFile := fmt.Sprintf("logs/%s/%s.log", email, currentDate)

	logData := domain.NewLog(time.Now(), dur.String(), email, operation, description, affectedFile)

	existingLogs, err := sm.FetchLogs(email)
	if err != nil && !pkg.IsNoSuchKeyError(err) {
		return err
	}

	logs := existingLogs[logFile]
	logs = append(logs, logData)

	byteData, err := json.Marshal(logs)
	if err != nil {
		return err
	}

	return sm.UploadToS3(logFile, bytes.NewReader(byteData))
}

func (sm *StorageManager) FetchLogs(email string) (map[string][]domain.Log, error) {
	prefix := fmt.Sprintf("logs/%s", email)

	files, err := sm.ListFiles(prefix)
	if err != nil {
		return nil, err
	}

	allLogs := make(map[string][]domain.Log)

	for _, file := range files {
		result, err := sm.DownloadFromS3(file)
		if err != nil && !pkg.IsNoSuchKeyError(err) {
			return nil, err
		}

		if result != nil {
			var logs []domain.Log
			if err := json.NewDecoder(result.Body).Decode(&logs); err != nil {
				return nil, err
			}
			result.Body.Close()
			allLogs[file] = logs
		}
	}

	return allLogs, nil
}
