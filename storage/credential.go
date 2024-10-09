package storage

import (
	"bytes"
	"dbhose/domain"
	"encoding/json"
	"fmt"
)

func (sm *StorageManager) StoreCreds(email string, creds domain.Credential) error {
	// download existing creds
	fileName := fmt.Sprintf("credentials/%s.json", email)
	result, err := sm.DownloadFromS3(fileName)
	if err != nil {
		return err
	}

	defer result.Body.Close()
	var existingCreds []domain.Credential
	if err := json.NewDecoder(result.Body).Decode(&existingCreds); err != nil {
		return err
	}

	// append new creds
	existingCreds = append(existingCreds, creds)

	// upload updated creds
	byteData, err := json.Marshal(existingCreds)
	if err != nil {
		return err
	}

	return sm.UploadToS3(fileName, bytes.NewReader(byteData))
}

func (sm *StorageManager) UpdateCreds(email string, creds domain.Credential) error {
	// download existing creds
	fileName := fmt.Sprintf("credentials/%s.json", email)
	result, err := sm.DownloadFromS3(fileName)
	if err != nil {
		return err
	}

	defer result.Body.Close()
	var existingCreds []domain.Credential
	if err := json.NewDecoder(result.Body).Decode(&existingCreds); err != nil {
		return err
	}

	// find and update creds
	for i, c := range existingCreds {
		if c.ID == creds.ID {
			existingCreds[i] = creds
			break
		}
	}

	// upload updated creds
	byteData, err := json.Marshal(existingCreds)
	if err != nil {
		return err
	}

	return sm.UploadToS3(fileName, bytes.NewReader(byteData))
}

func (sm *StorageManager) FindCredentialByID(email, id string) (domain.Credential, error) {
	fileName := fmt.Sprintf("credentials/%s.json", email)
	result, err := sm.DownloadFromS3(fileName)
	if err != nil {
		return domain.Credential{}, err
	}

	defer result.Body.Close()
	var creds []domain.Credential
	if err := json.NewDecoder(result.Body).Decode(&creds); err != nil {
		return domain.Credential{}, err
	}

	for _, c := range creds {
		if c.ID == id {
			return c, nil
		}
	}

	return domain.Credential{}, fmt.Errorf("credential with id %s not found", id)
}

func (sm *StorageManager) ListCredential(email string) ([]domain.Credential, error) {
	fileName := fmt.Sprintf("credentials/%s.json", email)
	result, err := sm.DownloadFromS3(fileName)
	if err != nil {
		return nil, err
	}

	defer result.Body.Close()

	var creds []domain.Credential
	if err := json.NewDecoder(result.Body).Decode(&creds); err != nil {
		return nil, err
	}

	return creds, nil
}

func (sm *StorageManager) DeleteCreds(email, id string) error {
	fileName := fmt.Sprintf("credentials/%s.json", email)
	result, err := sm.DownloadFromS3(fileName)
	if err != nil {
		return err
	}

	defer result.Body.Close()
	var existingCreds []domain.Credential
	if err := json.NewDecoder(result.Body).Decode(&existingCreds); err != nil {
		return err
	}

	// find and delete creds
	for i, c := range existingCreds {
		if c.ID == id {
			existingCreds = append(existingCreds[:i], existingCreds[i+1:]...)
			break
		}
	}

	// upload updated creds
	byteData, err := json.Marshal(existingCreds)
	if err != nil {
		return err
	}

	return sm.UploadToS3(fileName, bytes.NewReader(byteData))
}

func (sm *StorageManager) DeleteAllCreds(email string) error {
	fileName := fmt.Sprintf("credentials/%s.json", email)
	return sm.DeleteObject(fileName)
}
