package ui

import (
	"shorty/config"
	"shorty/types"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginUserPass(ctx fiber.Ctx) error {
	// Check if user/pass auth is configured
	if !IsUserPassConfigured() {
		return ctx.Status(fiber.StatusBadRequest).JSON(types.Response{
			Error:   true,
			Message: "Username/password authentication is not configured",
		})
	}

	var req LoginRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(types.Response{
			Error:   true,
			Message: "Invalid request body",
		})
	}

	// Validate credentials
	if req.Username != config.Use.App.Auth.User || req.Password != config.Use.App.Auth.Password {
		log.Warn().
			Str("username", req.Username).
			Msg("failed login attempt with incorrect credentials")
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: "Invalid username or password",
		})
	}

	// Get session
	sess, err := sessionStore.Get(ctx)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to get session")
		return ctx.Status(fiber.StatusInternalServerError).JSON(types.Response{
			Error:   true,
			Message: "Failed to initialize login session",
		})
	}
	defer sess.Release()

	if !sess.Fresh() {
		if err := sess.Regenerate(); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(types.Response{
				Error:   true,
				Message: "Failed to regenerate session",
			})
		}
	}

	// Set session data
	sess.Set("name", req.Username)
	if err := sess.Save(); err != nil {
		log.Error().Err(err).Msg("failed to save session")
		return ctx.Status(fiber.StatusInternalServerError).JSON(types.Response{
			Error:   true,
			Message: "Failed to create user session",
		})
	}

	return ctx.JSON(types.Response{
		Error:   false,
		Message: "Login successful",
	})
}
