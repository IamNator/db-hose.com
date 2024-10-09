package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	utils "dbhose/pkg"
)

func (sm *StorageManager) StoreBackup(email string, reader *bytes.Reader) error {
	hashedEmail := utils.Hash(email)
	timestamp := time.Now().UTC().Format("2006_01_02_15_04_05")
	fileKey := fmt.Sprintf("backups/%s/%s.sql", hashedEmail, timestamp)
	return sm.UploadToS3(fileKey, reader)
}

func (sm *StorageManager) FetchBackup(fileKey string) ([]byte, error) {
	result, err := sm.DownloadFromS3(fileKey)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	return io.ReadAll(result.Body)
}

func (sm *StorageManager) LogBackup(dur time.Duration, email, fileKey string) error {
	return sm.logOperation("Backup", dur, email, fileKey)
}

func (sm *StorageManager) LogRestore(dur time.Duration, email, fileKey string) error {
	return sm.logOperation("Restore", dur, email, fileKey)
}

func (sm *StorageManager) logOperation(operation string, dur time.Duration, email, fileKey string) error {
	hashedEmail := utils.Hash(email)
	currentDate := time.Now().UTC().Format("2006_01_02")
	logKey := fmt.Sprintf("logs/%s/%s.log", hashedEmail, currentDate)

	logData := map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"duration":  dur.String(),
		"email":     email,
		"event":     operation,
		"file":      fileKey,
	}

	existingLogs, err := sm.FetchLogs(email)
	if err != nil && !utils.IsNoSuchKeyError(err) {
		return err
	}

	logs := existingLogs[logKey]
	logs = append(logs, logData)

	byteData, err := json.Marshal(logs)
	if err != nil {
		return err
	}

	return sm.UploadToS3(logKey, bytes.NewReader(byteData))
}

func (sm *StorageManager) FetchLogs(email string) (map[string][]map[string]interface{}, error) {
	hashedEmail := utils.Hash(email)
	prefix := fmt.Sprintf("logs/%s", hashedEmail)

	files, err := sm.ListFiles(prefix)
	if err != nil {
		return nil, err
	}

	allLogs := make(map[string][]map[string]interface{})

	for _, file := range files {
		result, err := sm.DownloadFromS3(file)
		if err != nil && !utils.IsNoSuchKeyError(err) {
			return nil, err
		}

		if result != nil {
			var logs []map[string]interface{}
			if err := json.NewDecoder(result.Body).Decode(&logs); err != nil {
				return nil, err
			}
			result.Body.Close()
			allLogs[file] = logs
		}
	}

	return allLogs, nil
}
