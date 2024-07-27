package s3

import (
	"bytes"
	"dbhose/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func StoreUser(user models.User) error {
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

func GetUser(email string) (models.User, error) {
	key := fmt.Sprintf("users/%s.json", email)
	result, err := DownloadFromS3(bucket, key)
	if err != nil {
		return models.User{}, err
	}
	defer result.Body.Close()

	var user models.User
	if err := json.NewDecoder(result.Body).Decode(&user); err != nil {
		return models.User{}, err
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

func UpdateUser(user models.User) error {
	return StoreUser(user)
}
