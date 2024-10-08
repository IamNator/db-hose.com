package storage

import (
	"bytes"
	"dbhose/domain"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func StoreUser(user domain.User) error {
	user.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	key := fmt.Sprintf("users/%s.json", user.Email)
	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return UploadToS3(bucket, key, bytes.NewReader(userBytes))
}

func GetUser(email string) (domain.User, error) {
	key := fmt.Sprintf("users/%s.json", email)
	result, err := DownloadFromS3(bucket, key)
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

func DeleteUser(email string) error {
	key := fmt.Sprintf("users/%s.json", email)
	svc := s3.New(sess)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func UpdateUser(user domain.User) error {
	return StoreUser(user)
}
