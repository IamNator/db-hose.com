package storage

import (
	"dbhose/config"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type StorageManager struct {
	bucket string
	sess   *session.Session
}

func New() (*StorageManager, error) {
	bucket := config.GetEnvVar("S3_BUCKET_NAME")
	region := config.GetEnvVar("AWS_REGION")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	return &StorageManager{bucket: bucket, sess: sess}, nil
}

func (sm *StorageManager) UploadToS3(key string, body io.ReadSeeker, meta ...map[string]*string) error {
	svc := s3.New(sm.sess)

	input := &s3.PutObjectInput{
		Bucket: aws.String(sm.bucket),
		Key:    aws.String(key),
		Body:   body,
	}

	var metadata = make(map[string]*string)
	for _, m := range meta {
		for k, v := range m {
			metadata[k] = v
		}
	}

	if len(metadata) > 0 {
		input.SetMetadata(metadata)
	}

	_, err := svc.PutObject(input)

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
