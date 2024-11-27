package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioService struct {
	Client *minio.Client
}

// InitMinioClient initializes a MinIO client.
func InitMinioClient(endpoint, accessKeyID, secretAccessKey string) (*MinioService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false, // Use `true` for HTTPS, otherwise `false` for HTTP
	})
	if err != nil {
		log.Println("Error initializing MinIO client:", err)
		return nil, err
	}

	return &MinioService{Client: client}, nil
}

// CreateBucket creates a new bucket in MinIO if it does not already exist.
func (ms *MinioService) CreateBucket(bucketName string) error {
	ctx := context.Background()
	exists, err := ms.Client.BucketExists(ctx, bucketName)
	if err != nil {
		log.Printf("Error checking bucket existence: %v\n", err)
		return err
	}
	if !exists {
		err := ms.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Printf("Error creating bucket %s: %v\n", bucketName, err)
			return err
		}
		log.Printf("Bucket %s created successfully\n", bucketName)
	}
	return nil
}

// UploadFile uploads a file to the specified bucket.
func (ms *MinioService) UploadFile(bucketName, objectName string, file multipart.File, fileSize int64, contentType string) (string, error) {
	ctx := context.Background()

	// Ensure the bucket exists
	if err := ms.CreateBucket(bucketName); err != nil {
		return "", err
	}

	// Upload the file
	_, err := ms.Client.PutObject(ctx, bucketName, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		log.Printf("Error uploading file %s: %v\n", objectName, err)
		return "", err
	}

	log.Printf("File %s uploaded successfully to bucket %s\n", objectName, bucketName)
	return fmt.Sprintf("%s/%s", bucketName, objectName), nil
}

// GenerateUniqueFileName creates a unique file name using the current timestamp.
func GenerateUniqueFileName(originalFileName string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d_%s", timestamp, originalFileName)
}
