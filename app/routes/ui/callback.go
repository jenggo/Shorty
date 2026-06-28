package ui

import (
	"fmt"
	"shorty/config"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
	"github.com/rs/zerolog/log"
)

func fetchOAuthUser(accessToken string) (oauthUserResponse, error) {
	cc := client.New()
	cc.SetTimeout(10 * time.Second)
	resp, err := cc.Get(fmt.Sprintf("%s/api/v4/user", config.Use.Oauth.BaseURL), client.Config{
		Header: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", accessToken),
		},
	})
	if err != nil {
		return oauthUserResponse{}, fmt.Errorf("failed to fetch user information: %w", err)
	}

	if resp.StatusCode() != fiber.StatusOK {
		return oauthUserResponse{}, fmt.Errorf("failed to authenticate with GitLab: invalid response (status %d)", resp.StatusCode())
	}

	var user oauthUserResponse
	if err := json.Unmarshal(resp.Body(), &user); err != nil {
		return oauthUserResponse{}, fmt.Errorf("failed to decode user info: %w", err)
	}

	return user, nil
}

func Callback(ctx fiber.Ctx) error {
	basePath := ""
	routePath := ctx.Route().Path

	if strings.Contains(ctx.Path(), "/web/") {
		basePath = "/web"
	}

	sess, err := sessionStore.Get(ctx)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to get session")
		return ctx.Redirect().To(basePath + "/login?error=Failed to initialize session")
	}
	defer sess.Release()

	expectedState := sess.Get("oauth_state")
	if expectedState != ctx.Query("state") {
		return ctx.Redirect().To(basePath + "/login?error=Invalid OAuth state, possible CSRF attack")
	}

	code := ctx.Query("code")
	token, err := getOAuthConfig(routePath).Exchange(ctx.Context(), code)
	if err != nil {
		log.Error().Err(err).Msg("failed to exchange code for token")
		return ctx.Redirect().To(basePath + "/login?error=Failed to authenticate with GitLab")
	}

	user, err := fetchOAuthUser(token.AccessToken)
	if err != nil {
		log.Error().Err(err).Send()
		return ctx.Redirect().To(basePath + "/login?error=" + err.Error())
	}

	if user.External {
		log.Warn().Str("username", user.Username).Msg("external user attempted to login")
		return ctx.Redirect().To(basePath + "/login?error=External users are not allowed to login")
	}

	sess.Set("name", user.Username)
	if err := sess.Save(); err != nil {
		log.Error().Err(err).Msg("failed to save session")
		return ctx.Redirect().To(basePath + "/login?error=Failed to create user session")
	}

	return ctx.Redirect().To(basePath + "/")
}
