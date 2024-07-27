package s3

import (
	"bytes"
	"dbhose/utils"
	"io"
	"log"
	"os"
	"time"

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

func LogBackup(email, fileKey string) error {
	// Log the backup operation to S3
	logKey := "logs/" + email + ".log"
	logData := []byte(time.Now().Format("2006-01-02 15:04:05") + " Backup: " + fileKey + "\n")
	return appendLog(logKey, logData)
}

func LogRestore(email, fileKey string) error {
	// Log the restore operation to S3
	logKey := "logs/" + email + ".log"
	logData := []byte(time.Now().Format("2006-01-02 15:04:05") + " Restore: " + fileKey + "\n")
	return appendLog(logKey, logData)
}

func appendLog(logKey string, logData []byte) error {
	// Get the existing log data
	result, err := DownloadFromS3(bucket, logKey)
	if err != nil && !utils.IsNoSuchKeyError(err) {
		return err
	}

	var buf bytes.Buffer
	if result != nil {
		if _, err := io.Copy(&buf, result.Body); err != nil {
			return err
		}
		defer result.Body.Close()
	}

	// Append the new log data
	buf.Write(logData)

	// Upload the combined log data
	return UploadToS3(bucket, logKey, bytes.NewReader(buf.Bytes()))
}
