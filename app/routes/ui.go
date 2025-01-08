package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"shorty/config"
	"shorty/pkg"
	"shorty/types"
	"shorty/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

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
	state := utils.GenerateState()

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

func UIUpload(ctx *fiber.Ctx) error {
	if err := validateSession(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(types.Response{
			Error:   true,
			Message: "Invalid file upload",
		})
	}

	mfile, err := file.Open()
	if err != nil {
		return fmt.Errorf("Failed to process file %s: %v", file.Filename, err)
	}
	defer mfile.Close()

	s3, err := pkg.NewS3(ctx.Context())
	if err != nil {
		return err
	}

	fileName := utils.SlugifyFilename(file.Filename)

	if err := s3.Upload(ctx.Context(), fileName, mfile, file.Size); err != nil {
		return fmt.Errorf("Failed to upload %s", fileName)
	}

	url, err := s3.GeneratePresignedURL(ctx.Context(), fileName, config.Use.S3.Expired)
	if err != nil {
		return fmt.Errorf("Failed to generate presigned url: %v", err)
	}

	shorty := pkg.HumanFriendlyEnglishString(8)

	if err := pkg.Redis.Set(ctx.Context(), shorty, url, config.Use.S3.Expired, true); err != nil {
		return err
	}

	return ctx.JSON(types.Response{
		Error:   false,
		Message: fmt.Sprintf("%s/%s", ctx.BaseURL(), shorty),
	})
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
			"username":  name,
			"s3Enabled": config.Use.S3.Enable,
		},
	})
}
