package s3

import (
	"bytes"
	"dbhose/models"
	"dbhose/utils"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func StoreCreds(email string, creds models.Credentials) error {
	hashedEmail := utils.Hash(email)
	key := fmt.Sprintf("credentials/%s/%s.json", hashedEmail, creds.Key)
	credsBytes, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	return UploadToS3(bucket, key, bytes.NewReader(credsBytes))
}

func UpdateCreds(email string, creds models.Credentials) error {
	return StoreCreds(email, creds)
}

func GetCreds(email, credKey string) (models.Credentials, error) {
	hashedEmail := utils.Hash(email)
	key := fmt.Sprintf("credentials/%s/%s.json", hashedEmail, credKey)
	result, err := DownloadFromS3(bucket, key)
	if err != nil {
		return models.Credentials{}, err
	}
	defer result.Body.Close()

	var creds models.Credentials
	if err := json.NewDecoder(result.Body).Decode(&creds); err != nil {
		return models.Credentials{}, err
	}
	return creds, nil
}

func FindAllCreds(email string) ([]models.Credentials, error) {
	hashedEmail := utils.Hash(email)
	key := fmt.Sprintf("credentials/%s/", hashedEmail)
	svc := s3.New(sess)
	result, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	var creds []models.Credentials
	for _, obj := range result.Contents {
		result, err := DownloadFromS3(bucket, *obj.Key)
		if err != nil {
			return nil, err
		}
		defer result.Body.Close()

		var cred models.Credentials
		if err := json.NewDecoder(result.Body).Decode(&cred); err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	return creds, nil
}

func DeleteCreds(email, credKey string) error {
	hashedEmail := utils.Hash(email)
	key := fmt.Sprintf("credentials/%s/%s.json", hashedEmail, credKey)
	svc := s3.New(sess)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func DeleteAllCreds(email string) error {
	hashedEmail := utils.Hash(email)
	key := fmt.Sprintf("credentials/%s/", hashedEmail)
	svc := s3.New(sess)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}
