package ui

import (
	"fmt"
	"shorty/config"
	"shorty/types"
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
		return fmt.Errorf("failed to get session")
	}
	defer sess.Release()

	expectedState := sess.Get("oauth_state")
	if expectedState != ctx.Query("state") {
		return ctx.Status(fiber.StatusBadRequest).JSON(types.Response{
			Error:   true,
			Message: "Invalid OAuth state",
		})
	}

	code := ctx.Query("code")
	token, err := oauthConfig.Exchange(ctx.Context(), code)
	if err != nil {
		log.Error().Err(err).Msg("failed to exchange code for token")
		return fmt.Errorf("failed to exchange code for token")
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
		return fmt.Errorf("failed to create token request")
	}

	if resp.StatusCode() != fiber.StatusOK {
		log.Error().Caller().Int("status code", resp.StatusCode()).Msg("Failed to authenticate user")
		return fmt.Errorf("Failed to authenticate user")
	}

	var user oauthUserResponse
	if err := json.Unmarshal(resp.Body(), &user); err != nil {
		log.Error().Err(err).Msg("failed to decode user info")
		return fmt.Errorf("failed to decode user info")
	}

	// Check if user is external
	if user.External {
		log.Warn().
			Str("username", user.Username).
			Msg("external user attempted to login")
		return ctx.Status(fiber.StatusForbidden).JSON(types.Response{
			Error:   true,
			Message: "external users are not allowed to login",
		})
	}

	// Set session
	sess.Set("name", user.Username)
	if err := sess.Save(); err != nil {
		log.Error().Err(err).Msg("failed to save session")
		return fmt.Errorf("failed to save session")
	}

	return ctx.Redirect().To("/")
}
