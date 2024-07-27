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
	awsRegion = os.Getenv("AWS_REGION")
	bucket    = os.Getenv("S3_BUCKET_NAME")
	sess      *session.Session
)

func init() {
	var err error
	sess, err = session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
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
