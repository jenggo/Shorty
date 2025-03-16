package ui

import (
	"shorty/config"
	"shorty/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/minio"
	"github.com/gofiber/storage/redis/v3"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

var (
	sessionStore = session.NewStore()
	oauthConfig  *oauth2.Config
	// baseURL      string
)

type oauthUserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	External bool   `json:"external"`
}

func InitOAuth() {
	oauthConfig = &oauth2.Config{
		ClientID:     config.Use.Oauth.ClientID,
		ClientSecret: config.Use.Oauth.ClientSecret,
		Scopes:       []string{"read_user"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.Use.Oauth.BaseURL + "/oauth/authorize",
			TokenURL: config.Use.Oauth.BaseURL + "/oauth/token",
		},
	}
}

func getOAuthConfig(path string) *oauth2.Config {
	cfg := *oauthConfig
	cfg.RedirectURL = config.Use.App.BaseURL + "/auth/gitlab/callback"
	if strings.Contains(path, "/web/") {
		cfg.RedirectURL = config.Use.App.BaseURL + "/web/auth/gitlab/callback"
	}

	return &cfg
}

func InitStore() {
	redisPort, _ := strconv.Atoi(config.Use.Redis.Port)
	redisStore := redis.New(redis.Config{
		Host:     config.Use.Redis.Host,
		Port:     redisPort,
		Password: config.Use.Redis.Password,
		Database: config.Use.Redis.DB.Auth + 1,
	})

	sessionStore = session.NewStore(session.Config{
		Storage:         redisStore,
		AbsoluteTimeout: 168 * time.Hour,
		CookieSecure:    true,
		CookieHTTPOnly:  true,
	})

	utils.Storage = minio.New(minio.Config{
		Endpoint: config.Use.S3.Endpoint,
		Bucket:   config.Use.S3.Bucket,
		Secure:   true,
		Credentials: minio.Credentials{
			AccessKeyID:     config.Use.S3.Key.Access,
			SecretAccessKey: config.Use.S3.Key.Secret,
		},
	})

	if err := utils.Storage.CheckBucket(); err != nil {
		log.Fatal().Err(err).Send()
	}
}
