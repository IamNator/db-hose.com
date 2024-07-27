package s3

import (
	"bytes"
	"dbhose/utils"
	"encoding/json"
	"time"
)

func LogBackup(dur time.Duration, email, fileKey string) error {

	// Log the backup operation to S3
	currentDate := time.Now().UTC().Format("2006_01_02")
	logKey := "logs/" + email + "/" + currentDate + ".log"

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
	currentDate := time.Now().UTC().Format("2006_01_02")
	logKey := "logs/" + email + "/" + currentDate + ".log"

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
