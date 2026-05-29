package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"monostack/internal/domain"
)

type S3Adapter struct{}

var _ domain.S3Manager = (*S3Adapter)(nil)

func NewS3Adapter() *S3Adapter {
	return &S3Adapter{}
}

func (a *S3Adapter) ListBuckets(ctx context.Context, cfg *domain.AWSConfig) ([]domain.S3Bucket, error) {
	if cfg.UseMock {
		return []domain.S3Bucket{
			{Name: "mock-photos-bucket"},
			{Name: "mock-billing-records"},
			{Name: "mock-static-assets-local"},
		}, nil
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	out, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	var buckets []domain.S3Bucket
	for _, b := range out.Buckets {
		buckets = append(buckets, domain.S3Bucket{
			Name: aws.ToString(b.Name),
		})
	}
	return buckets, nil
}

func (a *S3Adapter) ListObjects(ctx context.Context, cfg *domain.AWSConfig, bucket string, prefix string) ([]domain.S3Object, error) {
	if cfg.UseMock {
		return []domain.S3Object{
			{Key: "images/logo.png", Size: 45600, LastModified: time.Now().Format(time.RFC3339)},
			{Key: "images/bg.jpg", Size: 120500, LastModified: time.Now().Format(time.RFC3339)},
			{Key: "index.html", Size: 1045, LastModified: time.Now().Format(time.RFC3339)},
			{Key: "notes.txt", Size: 245, LastModified: time.Now().Format(time.RFC3339)},
		}, nil
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var objects []domain.S3Object
	for _, o := range out.Contents {
		objects = append(objects, domain.S3Object{
			Key:          aws.ToString(o.Key),
			Size:         aws.ToInt64(o.Size),
			LastModified: aws.ToTime(o.LastModified).Format(time.RFC3339),
		})
	}
	return objects, nil
}

func (a *S3Adapter) DeleteObject(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string) error {
	if cfg.UseMock {
		return nil
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

func (a *S3Adapter) DeleteBucket(ctx context.Context, cfg *domain.AWSConfig, bucket string) error {
	if cfg.UseMock {
		return nil
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	_, err = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	return nil
}

func (a *S3Adapter) CreateBucket(ctx context.Context, cfg *domain.AWSConfig, name string) error {
	if cfg.UseMock {
		return nil
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

func (a *S3Adapter) CreateFolder(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string) error {
	if cfg.UseMock {
		return nil
	}

	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return fmt.Errorf("folder key is required")
	}
	if !strings.HasSuffix(trimmedKey, "/") {
		trimmedKey += "/"
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(trimmedKey),
		Body:   bytes.NewReader(nil),
	})
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}

func (a *S3Adapter) UploadObject(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string, filePath string) error {
	if cfg.UseMock {
		return nil
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	client := s3.NewFromConfig(awsCfg)
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

func (a *S3Adapter) GetPresignedURL(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string) (string, error) {
	if cfg.UseMock {
		return "https://mock-s3-presigned-url.com/" + bucket + "/" + key, nil
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	presignClient := s3.NewPresignClient(client)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", fmt.Errorf("failed to presign get object: %w", err)
	}

	return presignedReq.URL, nil
}

func (a *S3Adapter) DownloadObject(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string, destPath string) error {
	if cfg.UseMock {
		return os.WriteFile(destPath, []byte("Mock file content"), 0600)
	}

	awsCfg, err := GetSDKConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer out.Body.Close()

	cleanDest := filepath.Clean(destPath)
	if err := os.MkdirAll(filepath.Dir(cleanDest), 0750); err != nil {
		return fmt.Errorf("failed to create destination folder: %w", err)
	}

	file, err := os.OpenFile(cleanDest, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, out.Body)
	if err != nil {
		return fmt.Errorf("failed to save object to disk: %w", err)
	}

	return nil
}
