package storage

import (
	"bytes"
	"dbhose/domain"
	utils "dbhose/pkg"
	"encoding/json"
	"fmt"
)

func (sm *StorageManager) StoreCreds(email string, creds domain.Credential) error {
	hashedEmail := utils.Hash(email)
	key := fmt.Sprintf("credentials/%s/%s.json", hashedEmail, creds.Key)
	credsBytes, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	return sm.UploadToS3(key, bytes.NewReader(credsBytes))
}

func (sm *StorageManager) UpdateCreds(email string, creds domain.Credential) error {
	return sm.StoreCreds(email, creds)
}

func (sm *StorageManager) GetCreds(email, credKey string) (domain.Credential, error) {
	hashedEmail := utils.Hash(email)
	key := fmt.Sprintf("credentials/%s/%s.json", hashedEmail, credKey)
	result, err := sm.DownloadFromS3(key)
	if err != nil {
		return domain.Credential{}, err
	}
	defer result.Body.Close()

	var creds domain.Credential
	if err := json.NewDecoder(result.Body).Decode(&creds); err != nil {
		return domain.Credential{}, err
	}
	return creds, nil
}

func (sm *StorageManager) ListCreds(email string) ([]domain.Credential, error) {
	hashedEmail := utils.Hash(email)
	prefix := fmt.Sprintf("credentials/%s/", hashedEmail)
	files, err := sm.ListFiles(prefix)
	if err != nil {
		return nil, err
	}

	var creds []domain.Credential
	for _, file := range files {
		result, err := sm.DownloadFromS3(file)
		if err != nil {
			return nil, err
		}
		defer result.Body.Close()

		var cred domain.Credential
		if err := json.NewDecoder(result.Body).Decode(&cred); err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	return creds, nil
}

func (sm *StorageManager) DeleteCreds(email, credKey string) error {
	hashedEmail := utils.Hash(email)
	key := fmt.Sprintf("credentials/%s/%s.json", hashedEmail, credKey)
	return sm.DeleteObject(key)
}

func (sm *StorageManager) DeleteAllCreds(email string) error {
	hashedEmail := utils.Hash(email)
	prefix := fmt.Sprintf("credentials/%s/", hashedEmail)
	files, err := sm.ListFiles(prefix)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := sm.DeleteObject(file); err != nil {
			return err
		}
	}
	return nil
}
