package storage

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type StorageManager struct {
	bucket string
	sess   *session.Session
}

func NewStorageManager() (*StorageManager, error) {
	bucket := os.Getenv("S3_BUCKET_NAME")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return nil, err
	}
	return &StorageManager{bucket: bucket, sess: sess}, nil
}

func (sm *StorageManager) UploadToS3(key string, body io.ReadSeeker) error {
	svc := s3.New(sm.sess)
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(sm.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

func (sm *StorageManager) DownloadFromS3(key string) (*s3.GetObjectOutput, error) {
	svc := s3.New(sm.sess)
	return svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(sm.bucket),
		Key:    aws.String(key),
	})
}

func (sm *StorageManager) ListFiles(prefix string) ([]string, error) {
	svc := s3.New(sm.sess)
	result, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(sm.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	var files []string
	for _, item := range result.Contents {
		files = append(files, *item.Key)
	}
	return files, nil
}

func (sm *StorageManager) DeleteObject(key string) error {
	svc := s3.New(sm.sess)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(sm.bucket),
		Key:    aws.String(key),
	})
	return err
}
