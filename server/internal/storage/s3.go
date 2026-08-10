package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Client wraps a MinIO/S3 client bound to a single bucket used to store
// uploaded CSV files.
type S3Client struct {
	client *minio.Client
	bucket string
}

// NewS3ClientFromEnv connects to the S3-compatible endpoint described by
// S3_ENDPOINT / S3_ACCESS_KEY / S3_SECRET_KEY / S3_BUCKET / S3_USE_SSL and
// ensures the target bucket exists (MinIO does not create it automatically).
func NewS3ClientFromEnv() (*S3Client, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("S3_ENDPOINT is required")
	}

	accessKey := os.Getenv("S3_ACCESS_KEY")
	if accessKey == "" {
		return nil, fmt.Errorf("S3_ACCESS_KEY is required")
	}

	secretKey := os.Getenv("S3_SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("S3_SECRET_KEY is required")
	}

	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required")
	}

	useSSL := os.Getenv("S3_USE_SSL") == "1" || os.Getenv("S3_USE_SSL") == "true"

	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("S3クライアント初期化エラー: %w", err)
	}

	s := &S3Client{client: cli, bucket: bucket}

	if err := s.ensureBucket(context.Background()); err != nil {
		return nil, err
	}

	fmt.Println("📦 S3(MinIO)接続成功:", endpoint, "bucket:", bucket)

	return s, nil
}

func (s *S3Client) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("バケット確認エラー: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("バケット作成エラー: %w", err)
		}
	}
	return nil
}

// Bucket returns the bucket name CSV objects are stored under.
func (s *S3Client) Bucket() string {
	return s.bucket
}

// UploadCSV stores the given CSV bytes and returns the object key it was
// stored under, namespaced by user and timestamped to avoid collisions.
func (s *S3Client) UploadCSV(ctx context.Context, userID int64, filename string, data []byte) (string, error) {
	key := fmt.Sprintf("csv/%d/%d_%s", userID, time.Now().UnixNano(), path.Base(filename))

	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: "text/csv",
	})
	if err != nil {
		return "", fmt.Errorf("S3アップロードエラー: %w", err)
	}

	return key, nil
}

// DownloadCSV fetches the raw bytes of a previously uploaded CSV object.
func (s *S3Client) DownloadCSV(ctx context.Context, objectKey string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("S3取得エラー: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("S3読み込みエラー: %w", err)
	}

	return data, nil
}
