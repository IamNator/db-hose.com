package s3

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"dbhose/utils"
)

func StoreBackup(email string, reader *bytes.Reader) error {
	// Store the backup data to S3
	hashedEmail := utils.Hash(email)
	timestamp := time.Now().UTC().Format("2006_01_02_15_04_05")
	fileKey := "backups/" + hashedEmail + "/" + timestamp + ".sql"
	return UploadToS3(bucket, fileKey, reader)
}

func FetchBackup(fileKey string) ([]byte, error) {
	// Restore the backup data from S3
	result, err := DownloadFromS3(bucket, fileKey)
	if err != nil {
		return nil, err
	}

	defer result.Body.Close()
	return io.ReadAll(result.Body)
}

func LogBackup(dur time.Duration, email, fileKey string) error {

	// Log the backup operation to S3
	hashedEmail := utils.Hash(email)
	currentDate := time.Now().UTC().Format("2006_01_02")
	logKey := "logs/" + hashedEmail + "/" + currentDate + ".log"

	rawData := map[string]any{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"duration":  dur.String(),
		"email":     email,
		"event":     "Backup",
		"file":      fileKey,
	}
	return appendLog(logKey, rawData)
}

func LogRestore(dur time.Duration, email, fileKey string) error {
	// Log the restore operation to S3
	hashedEmail := utils.Hash(email)
	currentDate := time.Now().UTC().Format("2006_01_02")
	logKey := "logs/" + hashedEmail + "/" + currentDate + ".log"

	rawData := map[string]any{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"duration":  dur.String(),
		"email":     email,
		"event":     "Restore",
		"file":      fileKey,
	}
	return appendLog(logKey, rawData)
}

func appendLog(logKey string, logData map[string]any) error {
	// Get the existing log data
	result, err := DownloadFromS3(bucket, logKey)
	if err != nil && !utils.IsNoSuchKeyError(err) {
		return err
	}

	var logs []map[string]any
	if result != nil {
		defer result.Body.Close()
		if err := json.NewDecoder(result.Body).Decode(&logs); err != nil {
			return err
		}

		// Append the new log data
		logs = append(logs, logData)
	} else {
		logs = []map[string]any{logData}
	}

	byteData, err := json.Marshal(logs)
	if err != nil {
		return err
	}

	// Upload the combined log data
	return UploadToS3(bucket, logKey, bytes.NewReader(byteData))
}

func FetchLogs(email string) (map[string][]map[string]any, error) {

	hashedEmail := utils.Hash(email)
	logKey := "logs/" + hashedEmail

	// list all the logs files
	files, err := ListFiles(bucket, logKey)
	if err != nil {
		return nil, err
	}

	var allLogs map[string][]map[string]any

	for _, file := range files {

		// Get the existing log data
		result, err := DownloadFromS3(bucket, file)
		if err != nil && !utils.IsNoSuchKeyError(err) {
			return nil, err
		}

		if result != nil {
			var logs []map[string]any
			if err := json.NewDecoder(result.Body).Decode(&logs); err != nil {
				return nil, err
			}

			result.Body.Close()

			allLogs[file] = logs
		}
	}

	return allLogs, nil
}
