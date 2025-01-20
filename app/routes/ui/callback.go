package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"shorty/config"
	"shorty/types"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

func Callback(ctx fiber.Ctx) error {
	sess, err := sessionStore.Get(ctx)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to get session")
		return fmt.Errorf("failed to get session")
	}
	defer sess.Release()

	code := ctx.Query("code")
	if code == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(types.Response{
			Error:   true,
			Message: "No authorization code received",
		})
	}

	// Verify state
	expectedState := sess.Get("oauth_state")
	if expectedState != ctx.Query("state") {
		return ctx.Status(fiber.StatusBadRequest).JSON(types.Response{
			Error:   true,
			Message: "Invalid OAuth state",
		})
	}

	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	// Prepare token request
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {oauthConfig.RedirectURL},
	}

	req, err := http.NewRequest("POST", oauthConfig.Endpoint.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.Error().Err(err).Msg("failed to create token request")
		return fmt.Errorf("failed to create token request")
	}

	req.SetBasicAuth(oauthConfig.ClientID, oauthConfig.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("failed to execute token request")
		return fmt.Errorf("failed to execute token request")
	}
	defer resp.Body.Close()

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		log.Error().Err(err).Msg("failed to parse token response")
		return fmt.Errorf("failed to parse token response")
	}

	// Get user info
	userReq, err := http.NewRequest("GET", config.Use.Oauth.BaseURL+"/api/v4/user", nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to create user info request")
		return fmt.Errorf("failed to create user info request")
	}

	userReq.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)
	userResp, err := httpClient.Do(userReq)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user info")
		return fmt.Errorf("failed to get user info")
	}
	defer userResp.Body.Close()

	var user oauthUserResponse
	if err := json.NewDecoder(userResp.Body).Decode(&user); err != nil {
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
