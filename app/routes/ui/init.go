package ui

import (
	"strings"
	"time"

	"shorty/config"
	"shorty/pkg"
	"shorty/utils"

	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/minio"
	"github.com/gofiber/storage/valkey"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

var (
	sessionStore = session.NewStore()
	oauthConfig  *oauth2.Config
	// baseURL      string
)

type oauthUserResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	ID       int    `json:"id"`
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
	// Build the session-store client via pkg so it picks up shorty's connection
	// tuning (see pkg.redisKeepAlive). gofiber's driver does not expose Dialer in
	// its Config, but NewFromConnection lets us inject a client we configured
	// ourselves, so the session connections stop pinging at 1 Hz.
	sessionClient, err := pkg.NewClient(config.Use.Redis.DB.Auth + 1)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	redisStore := valkey.NewFromConnection(sessionClient)

	sessionStore = session.NewStore(session.Config{
		Storage:         redisStore,
		AbsoluteTimeout: 168 * time.Hour,
		CookieSecure:    true,
		CookieHTTPOnly:  true,
	})

	if config.Use.S3.Enable {
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
}

// IsOAuthConfigured checks if OAuth is properly configured
func IsOAuthConfigured() bool {
	return config.Use.Oauth.Enable &&
		config.Use.Oauth.ClientID != "" &&
		config.Use.Oauth.ClientSecret != "" &&
		config.Use.Oauth.BaseURL != ""
}

// IsUserPassConfigured checks if username/password auth is configured
func IsUserPassConfigured() bool {
	return config.Use.App.Auth.User != "" && config.Use.App.Auth.Password != ""
}

// GetAuthMethods returns the configured authentication methods
func GetAuthMethods() map[string]bool {
	return map[string]bool{
		"oauth":    IsOAuthConfigured(),
		"userpass": IsUserPassConfigured(),
	}
}
