package pkg

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"shorty/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type s3 struct {
	m *minio.Client
}

func NewS3(ctx context.Context) (*s3, error) {
	c, err := minio.New(config.Use.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Use.S3.Key.Access, config.Use.S3.Key.Secret, ""),
		Secure: true,
	})
	c.SetAppInfo(config.AppName, config.AppVersion)
	if config.Use.S3.Tracing {
		c.TraceOn(os.Stdout)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client: %v", err)
	}

	ctX, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	found, err := c.BucketExists(ctX, config.Use.S3.Bucket)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("bucket %s not found", config.Use.S3.Bucket)
	}

	return &s3{m: c}, nil
}

func (s3 *s3) Download(ctx context.Context, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	object, err := s3.m.GetObject(ctx, config.Use.S3.Bucket, objectName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to download object %s: %v", objectName, err)
	}

	return object, nil
}

func (s3 *s3) CheckObject(ctx context.Context, objectName, md5Hash string) (bool, error) {
	objectStat, err := s3.m.StatObject(ctx, config.Use.S3.Bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "AccessDenied" {
			return false, fmt.Errorf("access denied with endpoint %s", config.Use.S3.Endpoint)
		}
		if errResponse.Code == "NoSuchBucket" {
			return false, fmt.Errorf("bucket %s not found", config.Use.S3.Bucket)
		}
		if errResponse.Code == "InvalidBucketName" {
			return false, fmt.Errorf("invalid bucket name %s", config.Use.S3.Bucket)
		}
		if errResponse.Code == "NoSuchKey" {
			return false, fmt.Errorf("not found %s", objectName)
		}
		return false, err
	}

	if md5Hash != "" {
		if objectStat.ETag != md5Hash {
			return false, nil
		}
	}

	return true, nil
}

func (s3 *s3) Upload(ctx context.Context, objectName string, file io.Reader, size int64) error {
	expires := time.Now().UTC().Add(config.Use.S3.Expired)
	if _, err := s3.m.PutObject(ctx, config.Use.S3.Bucket, objectName, file, size, minio.PutObjectOptions{Expires: expires}); err != nil {
		return fmt.Errorf("failed to upload object %s: %v", objectName, err)
	}

	return nil
}

func (s3 *s3) GeneratePresignedURL(ctx context.Context, objectName string, expired time.Duration) (string, error) {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "inline")

	objectExpired := time.Duration(expired.Seconds()) * time.Second

	url, err := s3.m.PresignedGetObject(ctx, config.Use.S3.Bucket, objectName, objectExpired, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %v", err)
	}

	return url.String(), nil
}

func (s3 *s3) Delete(ctx context.Context, objectName string) error {
	if err := s3.m.RemoveObject(ctx, config.Use.S3.Bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete object %s: %v", objectName, err)
	}

	return nil
}
