package storage

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// GetFile will read in a file and return a file type,
// This will read from the file system
func GetFile(fileDir string) (*os.File, error) {
	file, err := os.Open(fileDir)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// NewSession Returns a new AWS session.  Uses your creds in ~/.aws
func NewSession(S3Region string) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(S3Region)})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// AddFileToS3 takes in a session, fileDir, s3_Bucket, and s3_Dir_Path
// In this case fileDir is the local file to read in and upload to S3; in this case we expect it to be local to the code
// S3 Dir Path is the directory within the S3 Bucket to write to
func AddFileToS3(s *session.Session, file *os.File, s3Bucket string, s3DirPath string) error {

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	key := fmt.Sprintf("%s%s", s3DirPath, file.Name())

	_, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s3Bucket),
		Key:                  aws.String(key),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	return err
}

//GetHashFromS3 will return the hash of the remote
func GetHashFromS3(s *session.Session, s3Bucket string, s3Key string) (string, error) {
	var err error
	var hash string

	svc := s3.New(s)

	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Key),
	})

	if err != nil {
		log.Println("Error retrieving Etag ", err)
	}
	hash = *resp.ETag
	return hash, err
}
