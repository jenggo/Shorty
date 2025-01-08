package routes

import (
	"shorty/config"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"golang.org/x/oauth2"
)

var (
	store       = session.New()
	oauthConfig *oauth2.Config
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
		RedirectURL:  config.Use.Oauth.RedirectURI,
		Scopes:       []string{"read_user"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   config.Use.Oauth.BaseURL + "/oauth/authorize",
			TokenURL:  config.Use.Oauth.BaseURL + "/oauth/token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
}

func InitStore() {
	redisPort, _ := strconv.Atoi(config.Use.Redis.Port)
	redisStore := redis.New(redis.Config{
		Host:     config.Use.Redis.Host,
		Port:     redisPort,
		Password: config.Use.Redis.Password,
		Database: config.Use.Redis.DB.Auth + 1,
	})

	store = session.New(session.Config{
		Storage:        redisStore,
		Expiration:     168 * time.Hour,
		CookieSecure:   true,
		CookieHTTPOnly: true,
	})
}
