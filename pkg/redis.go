package pkg

import (
	"context"
	"fmt"
	"mime"
	"net"
	"net/url"
	"path/filepath"
	"time"

	"shorty/config"
	"shorty/types"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type redis struct {
	client *goredis.Client
}

var Redis, RedisAuth *redis

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

func (r *redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, checkFirst ...bool) error {
	if ttl < 1 {
		ttl = 30 * time.Minute
	}

	if len(checkFirst) > 0 && checkFirst[0] {
		added, err := r.client.SAdd(ctx, "all_values", value).Result()
		if err != nil {
			return err
		}

		if added == 0 {
			return fmt.Errorf("%s already exists", value)
		}
	}

	return r.client.Set(ctx, key, value, ttl).Err()
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
		if iter.Val() == "all_values" {
			continue
		}

		url := r.client.Get(ctx, iter.Val()).Val()
		expired := r.client.TTL(ctx, iter.Val())
		datas = append(datas, types.Shorten{
			Url:     url,
			File:    getFile(url),
			Shorty:  iter.Val(),
			Expired: expired.Val(),
		})
	}

	err = iter.Err()

	return
}

func getFile(input string) string {
	u, _ := url.Parse(input)
	transform := filepath.Base(u.Path)
	ext := filepath.Ext(transform)
	test := mime.TypeByExtension(ext)
	if test != "" {
		return transform
	}

	return test
}

func (r *redis) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
