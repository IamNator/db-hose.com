package storage

import (
	"bytes"
	"dbhose/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func (sm *StorageManager) StoreBackup(email string, scriptReader *bytes.Reader) error {
	migration := domain.Migration{
		Email:     email,
		TimeStamp: time.Now().UTC(),
		Size:      int64(scriptReader.Len()),
		Meta:      make(map[string]string),
	}

	scriptFile := fmt.Sprintf("backups/%s", migration.File())

	if err := sm.UploadToS3(scriptFile, scriptReader); err != nil {
		return err
	}

	return sm.updateMigrationHistory(migration)
}

func (sm *StorageManager) updateMigrationHistory(migration domain.Migration) error {
	key := fmt.Sprintf("migrations/%s", migration.Email)

	//download existing migrations
	result, err := sm.DownloadFromS3(key)
	if err != nil {
		return err
	}

	defer result.Body.Close()

	var migrations []domain.Migration
	if err := json.NewDecoder(result.Body).Decode(&migrations); err != nil {
		return err
	}

	//append new migration
	migrations = append(migrations, migration)

	//upload updated migrations
	byteData, err := json.Marshal(migrations)
	if err != nil {
		return err
	}

	return sm.UploadToS3(key, bytes.NewReader(byteData))
}

func (sm *StorageManager) FetchBackup(email string, tsmp time.Time) ([]byte, error) {
	timestamp := tsmp.UTC().Format("2006_01_02_15_04_05")
	scriptFile := fmt.Sprintf("backups/%s/%s.sql", email, timestamp)

	result, err := sm.DownloadFromS3(scriptFile)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

func (sm *StorageManager) ListBackups(email string) ([]domain.Migration, error) {
	key := fmt.Sprintf("migrations/%s", email)

	//download existing migrations
	result, err := sm.DownloadFromS3(key)
	if err != nil {
		return nil, err
	}

	defer result.Body.Close()

	var migrations []domain.Migration
	if err := json.NewDecoder(result.Body).Decode(&migrations); err != nil {
		return nil, err
	}

	return migrations, nil
}
