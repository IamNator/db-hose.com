package storage

import (
	"bytes"
	"dbhose/domain"
	"encoding/json"
	"fmt"
	"time"
)

func (sm *StorageManager) StoreUser(user domain.User) error {
	user.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	key := fmt.Sprintf("users/%s.json", user.Email)
	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return sm.UploadToS3(key, bytes.NewReader(userBytes))
}

func (sm *StorageManager) GetUser(email string) (domain.User, error) {
	key := fmt.Sprintf("users/%s.json", email)
	result, err := sm.DownloadFromS3(key)
	if err != nil {
		return domain.User{}, err
	}
	defer result.Body.Close()

	var user domain.User
	if err := json.NewDecoder(result.Body).Decode(&user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (sm *StorageManager) DeleteUser(email string) error {
	key := fmt.Sprintf("users/%s.json", email)
	return sm.DeleteObject(key)
}

func (sm *StorageManager) UpdateUser(user domain.User) error {
	user.UpdatedAt = time.Now()
	return sm.StoreUser(user)
}
