package pkg

import (
	"context"
	"fmt"
	"mime"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"shorty/config"
	"shorty/types"
	"shorty/utils"

	"github.com/minio/minio-go/v7"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type redis struct {
	client *goredis.Client
}

var Redis, RedisAuth *redis

const (
	s3CachePrefix = "s3_exists:"
	s3CredPrefix  = "s3_cred:"
)

func NewRedis(useDB ...int) (*redis, error) {
	db := config.Use.Redis.DB.Main
	if len(useDB) > 0 {
		db = useDB[0]
	}

	addr := net.JoinHostPort(config.Use.Redis.Host, config.Use.Redis.Port)
	client := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: config.Use.Redis.Password,
		DB:       db,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		log.Error().Caller().Err(err).Send()
		return nil, err
	}

	return &redis{client: client}, nil
}

func (r *redis) Close() {
	if err := r.client.Close(); err != nil {
		log.Error().Caller().Err(err).Send()
	}
}

func (r *redis) Set(ctx context.Context, key string, value any, ttl time.Duration, checkFirst ...bool) error {
	if ttl < 1 {
		ttl = 30 * time.Minute
	}

	valueStr := fmt.Sprint(value)

	if len(checkFirst) > 0 && checkFirst[0] {
		exists, err := r.client.Exists(ctx, valueStr).Result()
		if err != nil {
			return err
		}

		if exists > 0 {
			return fmt.Errorf("%s already exists", value)
		}
	}

	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return err
	}

	if file := checkIsS3File(valueStr); file != "" {
		s3CacheKey := s3CachePrefix + key
		r.client.Set(ctx, s3CacheKey, file, ttl)
	}

	return nil
}

// SetWithS3Credentials sets a URL with associated S3 credentials
func (r *redis) SetWithS3Credentials(ctx context.Context, key string, value any, s3Creds types.S3Credentials, ttl time.Duration, checkFirst ...bool) error {
	// First set the main URL
	if err := r.Set(ctx, key, value, ttl, checkFirst...); err != nil {
		return err
	}

	// Then store the credentials in a separate key
	s3CredKey := s3CredPrefix + key
	if err := r.client.Set(ctx, s3CredKey, utils.ToJSON(s3Creds), ttl).Err(); err != nil {
		// If we fail to store credentials, clean up the main key
		r.client.Del(ctx, key)
		return err
	}

	return nil
}

// GetS3Credentials retrieves S3 credentials for a key if they exist
func (r *redis) GetS3Credentials(ctx context.Context, key string) (types.S3Credentials, error) {
	var creds types.S3Credentials

	s3CredKey := s3CredPrefix + key
	data, err := r.client.Get(ctx, s3CredKey).Bytes()
	if err != nil {
		return creds, err
	}

	if err := utils.FromJSON(data, &creds); err != nil {
		return creds, err
	}

	return creds, nil
}

func (r *redis) Get(ctx context.Context, key string) (string, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		err = fmt.Errorf("not found %s", key)
	}

	return string(data), err
}

func (r *redis) GetAll(ctx context.Context) (datas []types.Shorten, err error) {
	iter := r.client.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		if strings.HasPrefix(key, s3CachePrefix) || strings.HasPrefix(key, s3CredPrefix) {
			continue
		}

		url := r.client.Get(ctx, key).Val()
		s3CacheKey := s3CachePrefix + key

		file, err := r.client.Get(ctx, s3CacheKey).Result()
		if err == goredis.Nil {
			file = checkIsS3File(url)
			ttl := 20 * time.Minute
			if file != "" {
				ttl = r.client.TTL(ctx, key).Val()
			}
			r.client.Set(ctx, s3CacheKey, file, ttl)
		}

		expired := r.client.TTL(ctx, iter.Val())
		datas = append(datas, types.Shorten{
			Url:     url,
			File:    file,
			Shorty:  iter.Val(),
			Expired: expired.Val(),
		})
	}

	err = iter.Err()

	return
}

func checkIsS3File(input string) string {
	gf := getFile(input)
	if gf == "" {
		return ""
	}

	byteFile, err := utils.Storage.Get(gf)
	if err == nil || byteFile != nil {
		return ""
	}

	return gf
}

func getFile(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return ""
	}

	if u.Host != config.Use.S3.Endpoint {
		return ""
	}

	transform := filepath.Base(u.Path)
	if transform == "" || transform == "." {
		return ""
	}

	ext := filepath.Ext(transform)
	if ext != "" {
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			return transform
		}
	}

	return transform
}

func (r *redis) Del(ctx context.Context, key string) error {
	s3CacheKey := s3CachePrefix + key
	s3CredKey := s3CredPrefix + key
	_ = r.client.Del(ctx, s3CacheKey).Err()
	_ = r.client.Del(ctx, s3CredKey).Err()
	return r.client.Del(ctx, key).Err()
}

func (r *redis) StartCleanupScheduler() {
	ticker := time.NewTicker(config.Use.S3.CleanupInterval)
	go func() {
		defer ticker.Stop()
		if err := r.removeExpiredS3Objects(); err != nil {
			log.Error().Err(err).Msg("failed to cleanup expired objects")
		}

		// Then run on each tick
		for range ticker.C {
			if err := r.removeExpiredS3Objects(); err != nil {
				log.Error().Err(err).Msg("failed to cleanup expired objects")
			}
			log.Debug().Msg("completed scheduled cleanup of expired S3 objects")
		}
	}()
}

func (r *redis) removeExpiredS3Objects() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Map to store valid S3 files (both from s3_exists: keys and URL values)
	validS3Files := make(map[string]struct{})

	// First scan for all regular keys to find valid S3 URLs
	iter := r.client.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// Skip special prefix keys in this pass
		if strings.HasPrefix(key, s3CachePrefix) || strings.HasPrefix(key, s3CredPrefix) {
			continue
		}

		// Get URL value
		url, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		// Check if URL points to S3
		if file := getFile(url); file != "" {
			validS3Files[file] = struct{}{}
		}
	}

	// Now scan for s3_exists: keys to handle expired entries
	iter = r.client.Scan(ctx, 0, s3CachePrefix+"*", 0).Iterator()
	log.Debug().Dur("interval", config.Use.S3.CleanupInterval).Msg("starting scheduled cleanup of expired S3 objects")
	for iter.Next(ctx) {
		originalKey := strings.TrimPrefix(iter.Val(), s3CachePrefix)

		// Check if original key exists
		exists, err := r.client.Exists(ctx, originalKey).Result()
		if err != nil {
			log.Error().Caller().Err(err).Msg("failed to check key existence")
			continue
		}

		if exists == 0 {
			// Original key doesn't exist - clean up
			filename, err := r.client.Get(ctx, iter.Val()).Result()
			if err != nil {
				continue
			}

			// Skip if filename is empty
			if filename == "" {
				if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
					log.Error().Caller().Err(err).Msg("failed to delete empty s3_exists key")
				}
				continue
			}

			// Only delete if file is not referenced by any valid URL
			if _, stillValid := validS3Files[filename]; !stillValid {
				// Delete from S3
				if err := utils.Storage.Delete(filename); err != nil {
					log.Error().Caller().Err(err).Str("file", filename).Msg("failed to delete S3 object")
					continue
				}

				log.Info().Str("file", filename).Msg("removed expired S3 object")
			}

			// Delete the s3_exists: key
			if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
				log.Error().Caller().Err(err).Msg("failed to delete Redis key")
			}

			// // Also delete any s3_cred: key for this shortened URL
			// s3CredKey := s3CredPrefix + originalKey
			// if err := r.client.Del(ctx, s3CredKey).Err(); err != nil && err != goredis.Nil {
			// 	log.Error().Caller().Err(err).Msg("failed to delete S3 credentials key")
			// }
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to iterate Redis keys: %w", err)
	}

	// Check for orphaned S3 objects
	client := utils.Storage.Conn()
	objectCh := client.ListObjects(ctx, config.Use.S3.Bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			log.Error().Err(object.Err).Msg("error listing S3 objects")
			continue
		}

		// Only delete if object is not referenced by any valid URL
		if _, exists := validS3Files[object.Key]; !exists {
			// Delete the S3 object
			if err := utils.Storage.Delete(object.Key); err != nil {
				log.Error().Caller().Err(err).Str("file", object.Key).Msg("failed to delete orphaned S3 object")
				continue
			}

			// Find and delete any s3_exists: keys that reference this file
			iter := r.client.Scan(ctx, 0, s3CachePrefix+"*", 0).Iterator()
			for iter.Next(ctx) {
				filename, err := r.client.Get(ctx, iter.Val()).Result()
				if err != nil {
					continue
				}

				if filename == object.Key {
					if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
						log.Error().Caller().Err(err).Str("key", iter.Val()).Msg("failed to delete s3_exists key")
					}
				}
			}

			log.Info().Str("file", object.Key).Msg("removed orphaned S3 object and related keys")
		}
	}

	return nil
}
