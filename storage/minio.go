package storage

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client *minio.Client
	Bucket string
}

// NewMinioClient создает клиент для работы с MinIO
func NewMinioClient(endpoint, accessKey, secretKey, bucketName string) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // Установите `true`, если используете HTTPS
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing MinIO client: %w", err)
	}

	// Создать бакет, если его нет
	ctx := context.Background()
	err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("Bucket %s already exists", bucketName)
		} else {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinioClient{
		Client: client,
		Bucket: bucketName,
	}, nil
}

// UploadFile загружает файл в MinIO
func (mc *MinioClient) UploadFile(ctx context.Context, objectName, filePath, contentType string) (string, error) {
	info, err := mc.Client.FPutObject(ctx, mc.Bucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("error uploading file to MinIO: %w", err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", info.Key, info.Size)
	return fmt.Sprintf("/%s/%s", mc.Bucket, objectName), nil
}

// GetFileURL возвращает URL файла
func (mc *MinioClient) GetFileURL(objectName string) (string, error) {
	// Создать объект url.Values
	reqParams := url.Values{}
	reqParams.Set("response-content-disposition", "attachment; filename="+objectName)

	// Вызвать метод PresignedGetObject
	url, err := mc.Client.PresignedGetObject(context.Background(), mc.Bucket, objectName, time.Hour*24, reqParams)
	if err != nil {
		return "", fmt.Errorf("error generating file URL: %w", err)
	}
	return url.String(), nil
}
