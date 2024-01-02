package pkg

import (
	"context"
	"fmt"
	"net"
	"shorty/config"
	"time"

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

func (r *redis) Set(key string, value interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redis) Get(key string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	data, err := r.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		err = fmt.Errorf("not found %s", key)
	}

	return string(data), err
}

func (r *redis) Del(key string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return r.client.Del(ctx, key).Err()
}
