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
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"monostack/internal/domain"
	"monostack/internal/pkg/retry"
)

type S3Adapter struct{ cache *ClientCache }

var _ domain.S3Manager = (*S3Adapter)(nil)

func NewS3Adapter(cache *ClientCache) *S3Adapter {
	return &S3Adapter{cache: cache}
}

func (a *S3Adapter) ListBuckets(ctx context.Context, cfg *domain.AWSConfig) ([]domain.S3Bucket, error) {
	if cfg.UseMock {
		return []domain.S3Bucket{
			{Name: "mock-photos-bucket"},
			{Name: "mock-billing-records"},
			{Name: "mock-static-assets-local"},
		}, nil
	}

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 client: %w", err)
	}
	var out *s3.ListBucketsOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.ListBuckets(ctx, &s3.ListBucketsInput{})
		return innerErr
	})
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

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 client: %w", err)
	}
	var objects []domain.S3Object
	var continuationToken *string

	for {
		var out *s3.ListObjectsV2Output
		err = retry.Do(ctx, retry.DefaultConfig, func() error {
			var innerErr error
			out, innerErr = client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket:            aws.String(bucket),
				Prefix:            aws.String(prefix),
				ContinuationToken: continuationToken,
			})
			return innerErr
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, o := range out.Contents {
			objects = append(objects, domain.S3Object{
				Key:          aws.ToString(o.Key),
				Size:         aws.ToInt64(o.Size),
				LastModified: aws.ToTime(o.LastModified).Format(time.RFC3339),
			})
		}

		if out.NextContinuationToken == nil || *out.NextContinuationToken == "" {
			break
		}
		continuationToken = out.NextContinuationToken
	}
	return objects, nil
}

func (a *S3Adapter) DeleteObject(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		return innerErr
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

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
		return innerErr
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

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(name),
		})
		return innerErr
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

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
	}
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(trimmedKey),
			Body:   bytes.NewReader(nil),
		})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}

const multipartPartSize = 5 * 1024 * 1024   // 5 MB
const multipartMinThreshold = 5 * 1024 * 1024 // 5 MB

func (a *S3Adapter) UploadObject(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string, filePath string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
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

	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	if size >= multipartMinThreshold {
		return a.uploadMultipart(ctx, cfg, client, bucket, key, file, size)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	}

	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.PutObject(ctx, input)
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

func (a *S3Adapter) UploadObjectMultipart(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string, filePath string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
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

	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	if size < multipartMinThreshold {
		input := &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   file,
		}
		return retry.Do(ctx, retry.DefaultConfig, func() error {
			_, innerErr := client.PutObject(ctx, input)
			return innerErr
		})
	}

	return a.uploadMultipart(ctx, cfg, client, bucket, key, file, size)
}

func (a *S3Adapter) uploadMultipart(ctx context.Context, cfg *domain.AWSConfig, client *s3.Client, bucket, key string, file *os.File, size int64) error {
	createOut, err := client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to create multipart upload: %w", err)
	}
	uploadID := aws.ToString(createOut.UploadId)

	defer func() {
		if err != nil {
			_, _ = client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(bucket),
				Key:      aws.String(key),
				UploadId: createOut.UploadId,
			})
		}
	}()

	var parts []s3types.CompletedPart
	partNumber := int32(1)
	buf := make([]byte, multipartPartSize)

	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			reader := bytes.NewReader(buf[:n])
			uploadOut, uploadErr := client.UploadPart(ctx, &s3.UploadPartInput{
				Bucket:     aws.String(bucket),
				Key:        aws.String(key),
				UploadId:   aws.String(uploadID),
				PartNumber: aws.Int32(partNumber),
				Body:       reader,
			})
			if uploadErr != nil {
				err = fmt.Errorf("failed to upload part %d: %w", partNumber, uploadErr)
				return err
			}
			parts = append(parts, s3types.CompletedPart{
				PartNumber: aws.Int32(partNumber),
				ETag:       uploadOut.ETag,
			})
			partNumber++
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			err = fmt.Errorf("failed to read file: %w", readErr)
			return err
		}
	}

	_, err = client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}
	return nil
}

func (a *S3Adapter) UploadObjectWithMetadata(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string, filePath string, metadata map[string]string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
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

	input := &s3.PutObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		Body:     file,
		Metadata: metadata,
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(key), "."))
	if ct, ok := commonContentTypes[ext]; ok {
		input.ContentType = aws.String(ct)
	}

	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.PutObject(ctx, input)
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to upload object with metadata: %w", err)
	}
	return nil
}

func (a *S3Adapter) HeadObject(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string) (string, map[string]string, error) {
	if cfg.UseMock {
		return "", nil, nil
	}

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get S3 client: %w", err)
	}
	var out *s3.HeadObjectOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		return innerErr
	})
	if err != nil {
		return "", nil, err
	}

	metadata := make(map[string]string, len(out.Metadata))
	for k, v := range out.Metadata {
		metadata[k] = v
	}

	return aws.ToString(out.ContentType), metadata, nil
}

var commonContentTypes = map[string]string{
	"html":  "text/html",
	"css":   "text/css",
	"js":    "application/javascript",
	"json":  "application/json",
	"png":   "image/png",
	"jpg":   "image/jpeg",
	"jpeg":  "image/jpeg",
	"gif":   "image/gif",
	"svg":   "image/svg+xml",
	"pdf":   "application/pdf",
	"txt":   "text/plain",
	"xml":   "application/xml",
	"yaml":  "application/x-yaml",
	"yml":   "application/x-yaml",
	"csv":   "text/csv",
	"zip":   "application/zip",
	"gz":    "application/gzip",
}

func (a *S3Adapter) GetPresignedURL(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string) (string, error) {
	if cfg.UseMock {
		return "https://mock-s3-presigned-url.com/" + bucket + "/" + key, nil
	}

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get S3 client: %w", err)
	}
	presignClient := s3.NewPresignClient(client)

	var presignedReq *v4.PresignedHTTPRequest
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		presignedReq, innerErr = presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(15*time.Minute))
		return innerErr
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign get object: %w", err)
	}

	return presignedReq.URL, nil
}

func (a *S3Adapter) DownloadObject(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string, destPath string) error {
	if cfg.UseMock {
		return os.WriteFile(destPath, []byte("Mock file content"), 0600)
	}

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
	}
	var out *s3.GetObjectOutput
	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		var innerErr error
		out, innerErr = client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		return innerErr
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

func (a *S3Adapter) ListObjectVersions(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string) ([]domain.S3ObjectVersion, error) {
	if cfg.UseMock {
		return []domain.S3ObjectVersion{
			{Key: key, VersionID: "v-latest", IsLatest: true, Size: 45600, LastModified: time.Now().Format(time.RFC3339)},
			{Key: key, VersionID: "v-prev-001", IsLatest: false, Size: 45200, LastModified: time.Now().Add(-24 * time.Hour).Format(time.RFC3339)},
			{Key: key, VersionID: "v-prev-002", IsLatest: false, Size: 44800, LastModified: time.Now().Add(-48 * time.Hour).Format(time.RFC3339)},
		}, nil
	}

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 client: %w", err)
	}

	var versions []domain.S3ObjectVersion
	var keyMarker *string
	var versionIDMarker *string

	for {
		var out *s3.ListObjectVersionsOutput
		err = retry.Do(ctx, retry.DefaultConfig, func() error {
			var innerErr error
			out, innerErr = client.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
				Bucket:          aws.String(bucket),
				Prefix:          aws.String(key),
				KeyMarker:       keyMarker,
				VersionIdMarker: versionIDMarker,
			})
			return innerErr
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list object versions: %w", err)
		}

		for _, v := range out.Versions {
			if aws.ToString(v.Key) == key {
				versions = append(versions, domain.S3ObjectVersion{
					Key:          aws.ToString(v.Key),
					VersionID:    aws.ToString(v.VersionId),
					IsLatest:     aws.ToBool(v.IsLatest),
					Size:         aws.ToInt64(v.Size),
					LastModified: aws.ToTime(v.LastModified).Format(time.RFC3339),
				})
			}
		}

		if aws.ToBool(out.IsTruncated) {
			keyMarker = out.NextKeyMarker
			versionIDMarker = out.NextVersionIdMarker
		} else {
			break
		}
	}

	return versions, nil
}

func (a *S3Adapter) DeleteObjectVersion(ctx context.Context, cfg *domain.AWSConfig, bucket string, key string, versionID string) error {
	if cfg.UseMock {
		return nil
	}

	client, err := a.cache.S3(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get S3 client: %w", err)
	}

	err = retry.Do(ctx, retry.DefaultConfig, func() error {
		_, innerErr := client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket:    aws.String(bucket),
			Key:       aws.String(key),
			VersionId: aws.String(versionID),
		})
		return innerErr
	})
	if err != nil {
		return fmt.Errorf("failed to delete object version: %w", err)
	}
	return nil
}
