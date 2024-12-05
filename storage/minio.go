package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client *minio.Client
	Bucket string
}

// NewMinioClient создает клиент для работы с MinIO
func NewMinioClient() (*MinioClient, error) {
	// Получаем параметры из переменных окружения
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET_NAME")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // Установите `true`, если используете HTTPS
	})
	if err != nil {
		return nil, fmt.Errorf("error initializing MinIO client: %w", err)
	}

	// Проверяем, существует ли бакет
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	// Если бакет не существует, создаем его
	if !exists {
		err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Bucket %s created", bucketName)
	} else {
		log.Printf("Bucket %s already exists", bucketName)
	}

	return &MinioClient{
		Client: client,
		Bucket: bucketName,
	}, nil
}

// UploadFile загружает файл в MinIO
func (m *MinioClient) UploadFile(ctx context.Context, objectName, filePath, contentType string) (string, error) {
	info, err := m.Client.FPutObject(ctx, m.Bucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("error uploading file to MinIO: %w", err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", info.Key, info.Size)
	return fmt.Sprintf("/%s/%s", m.Bucket, objectName), nil
}

func (m *MinioClient) GeneratePresignedURL(objectName string, expires time.Duration) (string, error) {
	// Логируем параметры
	log.Printf("Generating presigned URL for bucket: %s, object: %s, expires in: %s", m.Bucket, objectName, expires)

	// Если путь начинается с system_images, возвращаем его без изменений
	if strings.HasPrefix(objectName, "system_images") {
		log.Printf("Path starts with 'system_images', returning original path: %s", objectName)
		return objectName, nil
	}

	// Создаем контекст
	ctx := context.Background()

	// Проверяем существование объекта
	exists, err := m.Client.StatObject(ctx, m.Bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		log.Printf("Error checking if object exists: %v", err)
		return "", fmt.Errorf("object %s does not exist in bucket %s: %w", objectName, m.Bucket, err)
	}
	log.Printf("Object exists: %+v", exists)

	// Генерация ссылки
	presignedURL, err := m.Client.PresignedGetObject(ctx, m.Bucket, objectName, expires, nil)
	if err != nil {
		log.Printf("Error generating presigned URL: %v", err)
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Выводим URL в логах для отладки
	log.Printf("Generated presigned URL: %s", presignedURL.String())
	return presignedURL.String(), nil
}
