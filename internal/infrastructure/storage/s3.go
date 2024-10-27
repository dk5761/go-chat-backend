package storage

import (
	"bytes"
	"context"
	"mime/multipart"
	"path/filepath"

	"github.com/dk5761/go-serv/configs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3StorageService struct {
	s3         *s3.S3
	bucketName string
}

func NewS3StorageService(cfg configs.S3Config) *S3StorageService {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	}))
	s3Client := s3.New(sess)
	return &S3StorageService{
		s3:         s3Client,
		bucketName: cfg.BucketName,
	}
}

func (s *S3StorageService) UploadFile(ctx context.Context, file multipart.File, fileName string) (string, error) {
	defer file.Close()
	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(file); err != nil {
		return "", err
	}

	contentType := getContentType(fileName)

	_, err := s.s3.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}

	fileURL := "https://" + s.bucketName + ".s3.amazonaws.com/" + fileName
	return fileURL, nil
}

func getContentType(fileName string) string {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}
