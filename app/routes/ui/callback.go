package ui

import (
	"fmt"
	"shorty/config"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
	"github.com/rs/zerolog/log"
)

func Callback(ctx fiber.Ctx) error {
	sess, err := sessionStore.Get(ctx)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to get session")
		return ctx.Redirect().To("/login?error=Failed to initialize session")
	}
	defer sess.Release()

	expectedState := sess.Get("oauth_state")
	if expectedState != ctx.Query("state") {
		return ctx.Redirect().To("/login?error=Invalid OAuth state, possible CSRF attack")
	}

	code := ctx.Query("code")
	token, err := oauthConfig.Exchange(ctx.Context(), code)
	if err != nil {
		log.Error().Err(err).Msg("failed to exchange code for token")
		return ctx.Redirect().To("/login?error=Failed to authenticate with GitLab")
	}

	cc := client.New()
	cc.SetTimeout(10 * time.Second)
	resp, err := cc.Get(fmt.Sprintf("%s/api/v4/user", config.Use.Oauth.BaseURL), client.Config{
		Header: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token.AccessToken),
		},
	})
	if err != nil {
		log.Error().Err(err).Send()
		return ctx.Redirect().To("/login?error=Failed to fetch user information")
	}

	if resp.StatusCode() != fiber.StatusOK {
		log.Error().Caller().Int("status code", resp.StatusCode()).Msg("Failed to authenticate user")
		return ctx.Redirect().To("/login?error=Failed to authenticate with GitLab: invalid response")
	}

	var user oauthUserResponse
	if err := json.Unmarshal(resp.Body(), &user); err != nil {
		log.Error().Err(err).Msg("failed to decode user info")
		return ctx.Redirect().To("/login?error=Failed to process user information")
	}

	// Check if user is external
	if user.External {
		log.Warn().
			Str("username", user.Username).
			Msg("external user attempted to login")
		return ctx.Redirect().To("/login?error=External users are not allowed to login")
	}

	// Set session
	sess.Set("name", user.Username)
	if err := sess.Save(); err != nil {
		log.Error().Err(err).Msg("failed to save session")
		return ctx.Redirect().To("/login?error=Failed to create user session")
	}

	return ctx.Redirect().To("/")
}
