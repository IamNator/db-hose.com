package storage

import (
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	bucket = os.Getenv("S3_BUCKET_NAME")
	sess   *session.Session
)

func Init() {

	bucket = os.Getenv("S3_BUCKET_NAME")

	var err error
	sess, err = session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		log.Fatalf("failed to create session, %v", err)
	}
}

func UploadToS3(bucket, key string, body io.ReadSeeker) error {
	svc := s3.New(sess)
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

func DownloadFromS3(bucket, key string) (*s3.GetObjectOutput, error) {
	svc := s3.New(sess)
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return result, err
}

func ListFiles(bucket, key string) ([]string, error) {
	svc := s3.New(sess)
	result, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
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
