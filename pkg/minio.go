package pkg

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"shorty/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

type s3 struct {
	m *minio.Client
}

func NewS3(ctx context.Context) (*s3, error) {
	c, err := minio.New(config.Use.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Use.S3.Key.Access, config.Use.S3.Key.Secret, ""),
		Secure: true,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
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

func (s3 *s3) Upload(ctx context.Context, objectName string, file io.Reader, size int64, expired time.Duration) error {
	if _, err := s3.m.StatObject(ctx, config.Use.S3.Bucket, objectName, minio.GetObjectOptions{}); err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "AccessDenied" {
			return fmt.Errorf("access denied with endpoint %s", config.Use.S3.Endpoint)
		}

		if errResponse.Code == "NoSuchBucket" || errResponse.Code == "InvalidBucketName" {
			return fmt.Errorf("bucket %s not found", config.Use.S3.Bucket)
		}

		if errResponse.Code == "NoSuchKey" {
			expires := time.Now().Local().Add(expired)
			if _, err := s3.m.PutObject(ctx, config.Use.S3.Bucket, objectName, file, size, minio.PutObjectOptions{
				Expires:               expires,
				ConcurrentStreamParts: true,
			}); err != nil {
				if err := s3.m.RemoveIncompleteUpload(ctx, config.Use.S3.Bucket, objectName); err != nil {
					log.Error().Caller().Err(err).Msgf("failed to remove incomplete upload %s", objectName)
				}

				return fmt.Errorf("failed to upload object %s: %v", objectName, err)
			}

			return nil
		}

		return err
	}

	return fmt.Errorf("%s already exists", objectName)
}

func (s3 *s3) GeneratePresignedURL(ctx context.Context, objectName string, expired time.Duration) (string, error) {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "inline")

	url, err := s3.m.PresignedGetObject(ctx, config.Use.S3.Bucket, objectName, expired, reqParams)
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
