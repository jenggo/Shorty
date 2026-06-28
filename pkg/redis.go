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
	"github.com/rs/zerolog/log"
	"github.com/valkey-io/valkey-go"
)

type redis struct {
	client valkey.Client
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

	opt := valkey.ClientOption{
		InitAddress: []string{addr},
		SelectDB:    db,
	}

	if config.Use.Redis.Password != "" {
		opt.Password = config.Use.Redis.Password
	}

	client, err := valkey.NewClient(opt)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return nil, err
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		log.Error().Caller().Err(err).Send()
		client.Close()
		return nil, err
	}

	return &redis{client: client}, nil
}

func (r *redis) Close() {
	r.client.Close()
}

func (r *redis) Set(ctx context.Context, key string, value any, ttl time.Duration, checkFirst ...bool) error {
	valueStr := fmt.Sprint(value)

	if len(checkFirst) > 0 && checkFirst[0] {
		resp := r.client.Do(ctx, r.client.B().Exists().Key(valueStr).Build())
		if err := resp.Error(); err != nil {
			return err
		}
		exists, _ := resp.AsInt64()
		if exists > 0 {
			return fmt.Errorf("%s already exists", value)
		}
	}

	var setCmdBuilder interface {
		Build() valkey.Completed
	}

	if ttl > 0 {
		setCmdBuilder = r.client.B().Set().Key(key).Value(fmt.Sprint(value)).ExSeconds(int64(ttl.Seconds()))
	} else {
		setCmdBuilder = r.client.B().Set().Key(key).Value(fmt.Sprint(value))
	}

	if err := r.client.Do(ctx, setCmdBuilder.Build()).Error(); err != nil {
		return err
	}

	if file := checkIsS3File(valueStr); file != "" {
		s3CacheKey := s3CachePrefix + key
		var cacheCmdBuilder interface {
			Build() valkey.Completed
		}

		if ttl > 0 {
			cacheCmdBuilder = r.client.B().Set().Key(s3CacheKey).Value(file).ExSeconds(int64(ttl.Seconds()))
		} else {
			cacheCmdBuilder = r.client.B().Set().Key(s3CacheKey).Value(file)
		}
		r.client.Do(ctx, cacheCmdBuilder.Build())
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
	credJSON := string(utils.ToJSON(s3Creds))

	var credCmdBuilder interface {
		Build() valkey.Completed
	}

	if ttl > 0 {
		credCmdBuilder = r.client.B().Set().Key(s3CredKey).Value(credJSON).ExSeconds(int64(ttl.Seconds()))
	} else {
		credCmdBuilder = r.client.B().Set().Key(s3CredKey).Value(credJSON)
	}

	if err := r.client.Do(ctx, credCmdBuilder.Build()).Error(); err != nil {
		// If we fail to store credentials, clean up the main key
		r.client.Do(ctx, r.client.B().Del().Key(key).Build())
		return err
	}

	return nil
}

// GetS3Credentials retrieves S3 credentials for a key if they exist
func (r *redis) GetS3Credentials(ctx context.Context, key string) (types.S3Credentials, error) {
	var creds types.S3Credentials

	s3CredKey := s3CredPrefix + key
	resp := r.client.Do(ctx, r.client.B().Get().Key(s3CredKey).Build())

	if err := resp.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return creds, fmt.Errorf("not found %s", s3CredKey)
		}
		return creds, err
	}

	data, err := resp.AsBytes()
	if err != nil {
		return creds, err
	}

	if err := utils.FromJSON(data, &creds); err != nil {
		return creds, err
	}

	return creds, nil
}

func (r *redis) Get(ctx context.Context, key string) (string, error) {
	resp := r.client.Do(ctx, r.client.B().Get().Key(key).Build())

	if err := resp.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return "", fmt.Errorf("not found %s", key)
		}
		return "", err
	}

	return resp.ToString()
}

func (r *redis) processScannedKey(ctx context.Context, key string) (types.Shorten, bool) {
	if strings.HasPrefix(key, s3CachePrefix) || strings.HasPrefix(key, s3CredPrefix) {
		return types.Shorten{}, false
	}

	urlResp := r.client.Do(ctx, r.client.B().Get().Key(key).Build())
	url, _ := urlResp.ToString()

	s3CacheKey := s3CachePrefix + key
	fileResp := r.client.Do(ctx, r.client.B().Get().Key(s3CacheKey).Build())

	var file string
	if valkey.IsValkeyNil(fileResp.Error()) {
		file = checkIsS3File(url)
		ttl := 20 * time.Minute
		if file != "" {
			ttlResp := r.client.Do(ctx, r.client.B().Ttl().Key(key).Build())
			if ttlSec, err := ttlResp.AsInt64(); err == nil && ttlSec > 0 {
				ttl = time.Duration(ttlSec) * time.Second
			}
		}
		r.client.Do(ctx, r.client.B().Set().Key(s3CacheKey).Value(file).ExSeconds(int64(ttl.Seconds())).Build())
	} else {
		file, _ = fileResp.ToString()
	}

	ttlResp := r.client.Do(ctx, r.client.B().Ttl().Key(key).Build())
	ttlSec, _ := ttlResp.AsInt64()

	return types.Shorten{
		Url:     url,
		File:    file,
		Shorty:  key,
		Expired: time.Duration(ttlSec) * time.Second,
	}, true
}

func (r *redis) GetAll(ctx context.Context) (datas []types.Shorten, err error) {
	var cursor uint64

	for {
		resp := r.client.Do(ctx, r.client.B().Scan().Cursor(cursor).Match("*").Build())
		if err := resp.Error(); err != nil {
			return nil, err
		}

		scanResp, err := resp.AsScanEntry()
		if err != nil {
			return nil, err
		}

		for _, key := range scanResp.Elements {
			if entry, ok := r.processScannedKey(ctx, key); ok {
				datas = append(datas, entry)
			}
		}

		cursor = scanResp.Cursor
		if cursor == 0 {
			break
		}
	}

	return datas, nil
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

	// Delete all related keys
	r.client.Do(ctx, r.client.B().Del().Key(s3CacheKey).Build())
	r.client.Do(ctx, r.client.B().Del().Key(s3CredKey).Build())

	return r.client.Do(ctx, r.client.B().Del().Key(key).Build()).Error()
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

func (r *redis) collectValidS3Files(ctx context.Context) (map[string]struct{}, error) {
	validS3Files := make(map[string]struct{})
	var cursor uint64

	for {
		resp := r.client.Do(ctx, r.client.B().Scan().Cursor(cursor).Match("*").Build())
		if err := resp.Error(); err != nil {
			return nil, err
		}

		scanResp, err := resp.AsScanEntry()
		if err != nil {
			return nil, err
		}

		for _, key := range scanResp.Elements {
			if strings.HasPrefix(key, s3CachePrefix) || strings.HasPrefix(key, s3CredPrefix) {
				continue
			}

			urlResp := r.client.Do(ctx, r.client.B().Get().Key(key).Build())
			if urlResp.Error() != nil {
				continue
			}

			url, _ := urlResp.ToString()
			if file := getFile(url); file != "" {
				validS3Files[file] = struct{}{}
			}
		}

		cursor = scanResp.Cursor
		if cursor == 0 {
			break
		}
	}

	return validS3Files, nil
}

func (r *redis) deleteCacheEntry(ctx context.Context, cacheKey string, validS3Files map[string]struct{}) {
	originalKey := strings.TrimPrefix(cacheKey, s3CachePrefix)

	existsResp := r.client.Do(ctx, r.client.B().Exists().Key(originalKey).Build())
	if existsResp.Error() != nil {
		log.Error().Caller().Err(existsResp.Error()).Msg("failed to check key existence")
		return
	}

	exists, _ := existsResp.AsInt64()
	if exists != 0 {
		return
	}

	filenameResp := r.client.Do(ctx, r.client.B().Get().Key(cacheKey).Build())
	if filenameResp.Error() != nil {
		return
	}

	filename, _ := filenameResp.ToString()
	if filename == "" {
		if err := r.client.Do(ctx, r.client.B().Del().Key(cacheKey).Build()).Error(); err != nil {
			log.Error().Caller().Err(err).Msg("failed to delete empty s3_exists key")
		}
		return
	}

	if _, stillValid := validS3Files[filename]; !stillValid {
		if err := utils.Storage.Delete(filename); err != nil {
			log.Error().Caller().Err(err).Str("file", filename).Msg("failed to delete S3 object")
			return
		}
		log.Info().Str("file", filename).Msg("removed expired S3 object")
	}

	if err := r.client.Do(ctx, r.client.B().Del().Key(cacheKey).Build()).Error(); err != nil {
		log.Error().Caller().Err(err).Msg("failed to delete Redis key")
	}
}

func (r *redis) cleanupExpiredCacheEntries(ctx context.Context, validS3Files map[string]struct{}) error {
	var cursor uint64

	for {
		resp := r.client.Do(ctx, r.client.B().Scan().Cursor(cursor).Match(s3CachePrefix+"*").Build())
		if err := resp.Error(); err != nil {
			return err
		}

		scanResp, err := resp.AsScanEntry()
		if err != nil {
			return err
		}

		for _, cacheKey := range scanResp.Elements {
			r.deleteCacheEntry(ctx, cacheKey, validS3Files)
		}

		cursor = scanResp.Cursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (r *redis) deleteOrphanedS3Keys(ctx context.Context, fileKey string) {
	var scanCursor uint64
	for {
		scanResp := r.client.Do(ctx, r.client.B().Scan().Cursor(scanCursor).Match(s3CachePrefix+"*").Build())
		if scanResp.Error() != nil {
			break
		}

		entry, err := scanResp.AsScanEntry()
		if err != nil {
			break
		}

		for _, key := range entry.Elements {
			filenameResp := r.client.Do(ctx, r.client.B().Get().Key(key).Build())
			if filenameResp.Error() != nil {
				continue
			}

			filename, _ := filenameResp.ToString()
			if filename == fileKey {
				if err := r.client.Do(ctx, r.client.B().Del().Key(key).Build()).Error(); err != nil {
					log.Error().Caller().Err(err).Str("key", key).Msg("failed to delete s3_exists key")
				}
			}
		}

		scanCursor = entry.Cursor
		if scanCursor == 0 {
			break
		}
	}
}

func (r *redis) cleanupOrphanedS3Objects(ctx context.Context, validS3Files map[string]struct{}) {
	client := utils.Storage.Conn()
	objectCh := client.ListObjects(ctx, config.Use.S3.Bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			log.Error().Err(object.Err).Msg("error listing S3 objects")
			continue
		}

		if _, exists := validS3Files[object.Key]; exists {
			continue
		}

		if err := utils.Storage.Delete(object.Key); err != nil {
			log.Error().Caller().Err(err).Str("file", object.Key).Msg("failed to delete orphaned S3 object")
			continue
		}

		r.deleteOrphanedS3Keys(ctx, object.Key)
		log.Info().Str("file", object.Key).Msg("removed orphaned S3 object and related keys")
	}
}

func (r *redis) removeExpiredS3Objects() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	validS3Files, err := r.collectValidS3Files(ctx)
	if err != nil {
		return err
	}

	log.Debug().Dur("interval", config.Use.S3.CleanupInterval).Msg("starting scheduled cleanup of expired S3 objects")

	if err := r.cleanupExpiredCacheEntries(ctx, validS3Files); err != nil {
		return err
	}

	r.cleanupOrphanedS3Objects(ctx, validS3Files)

	return nil
}
