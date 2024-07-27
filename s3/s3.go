package s3

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
