package routes

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"shorty/config"
	"shorty/types"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"github.com/rs/zerolog/log"
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

func generateState() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func getSession(ctx *fiber.Ctx) (*session.Session, error) {
	sess, err := store.Get(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get session")
		return nil, err
	}

	return sess, nil
}

func validateSession(ctx *fiber.Ctx) error {
	sess, err := getSession(ctx)
	if err != nil {
		return err
	}

	name := sess.Get("name")
	if name == nil {
		return errors.New("unauthorized access")
	}

	return nil
}

func UILogout(ctx *fiber.Ctx) error {
	sess, err := getSession(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get session")
		return fmt.Errorf("failed to get session")
	}

	if err := sess.Destroy(); err != nil {
		log.Error().Err(err).Msg("failed to destroy session")
		return fmt.Errorf("failed to logout")
	}

	return ctx.Render("logout", nil)
}

func UICreate(ctx *fiber.Ctx) error {
	if err := validateSession(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	return Shorten(ctx)
}

func UIDelete(ctx *fiber.Ctx) error {
	if err := validateSession(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	return Delete(ctx)
}

func UIChange(ctx *fiber.Ctx) error {
	if err := validateSession(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	return Change(ctx)
}

func UIOauthLogin(ctx *fiber.Ctx) error {
	state := generateState()

	sess, err := getSession(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get session")
		return fmt.Errorf("failed to get session")
	}

	_ = sess.Destroy()

	sess.Set("oauth_state", state)
	if err := sess.Save(); err != nil {
		log.Error().Err(err).Msg("failed to save session")
		return fmt.Errorf("failed to save session")
	}

	url := oauthConfig.AuthCodeURL(state)
	return ctx.JSON(types.Response{
		Error:   false,
		Message: url,
	})
}

func UICallback(ctx *fiber.Ctx) error {
	sess, err := getSession(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get session")
		return fmt.Errorf("failed to get session")
	}

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

	return ctx.Redirect("/")
}

func CheckSession(ctx *fiber.Ctx) error {
	sess, err := getSession(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: "No valid session",
		})
	}

	name := sess.Get("name")
	if name == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: "Not logged in",
		})
	}

	return ctx.JSON(fiber.Map{
		"error": false,
		"data": fiber.Map{
			"username": name,
		},
	})
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
		Expiration:     24 * time.Hour,
		CookieSecure:   true,
		CookieHTTPOnly: true,
	})
}
