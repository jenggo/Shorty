package app

import (
	"context"
	"fmt"
	"time"

	"shorty/config"
	"shorty/pkg"
	"shorty/types"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/keyauth"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// rateLimiter builds a per-minute, per-IP fixed-window limiter. scope must be
// unique per route group: limiter's default key is the bare client IP, so
// without a prefix every limiter sharing the storage would share one counter
// per IP.
func rateLimiter(storage fiber.Storage, scope string, limit int) fiber.Handler {
	return limiter.New(limiter.Config{
		Storage:           storage,
		LimiterMiddleware: limiter.FixedWindow{},
		Max:               limit,
		Expiration:        time.Minute,
		KeyGenerator:      func(c fiber.Ctx) string { return scope + ":" + c.IP() },
		LimitReached:      rateLimitReached,
	})
}

func rateLimitReached(c fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(types.Response{
		Error:   true,
		Message: "too many requests",
	})
}

// customBearerExtractor extracts Bearer tokens allowing $ character (for bcrypt hashes)
func customBearerExtractor(c fiber.Ctx) (string, error) {
	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return "", extractors.ErrNotFound
	}

	const bearerScheme = "Bearer "
	if len(authHeader) <= len(bearerScheme) || authHeader[:len(bearerScheme)] != bearerScheme {
		return "", extractors.ErrNotFound
	}

	token := authHeader[len(bearerScheme):]
	if token == "" {
		return "", extractors.ErrNotFound
	}

	return token, nil
}

func verifyKey() func(c fiber.Ctx) error {
	return keyauth.New(keyauth.Config{
		Extractor: extractors.FromCustom("Bearer", customBearerExtractor),
		Validator: func(c fiber.Ctx, key string) (bool, error) {
			if err := verifyAPIKey(c.Context(), key); err != nil {
				log.Error().Caller().Err(err).Str("path", c.Path()).Str("hash", key).
					Send()
				return false, keyauth.ErrMissingOrMalformedAPIKey
			}

			log.Info().Str("path", c.Path()).Msg("API key verified")

			return true, nil
		},
		ErrorHandler: keyAuthErrorHandler,
	})
}

func keyAuthErrorHandler(c fiber.Ctx, err error) error {
	ua := c.Get(fiber.HeaderUserAgent)
	ip := c.IP()
	method := c.Method()
	path := c.Path()

	// Extract Authorization header to see what was sent
	authHeader := c.Get(fiber.HeaderAuthorization)

	// Log with detailed information
	log.Error().Caller().Str("error_msg", err.Error()).Str("UserAgent", ua).Str("IP", ip).
		Str("Method", method).
		Str("Path", path).
		Str("Authorization_header", authHeader).
		Send()

	return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
		"error":   true,
		"message": err.Error(),
	})
}

func verifyAPIKey(ctx context.Context, clientBcryptHash string) error {
	// Check if it's already used (prevent replay attacks)
	if _, err := pkg.RedisAuth.Get(ctx, clientBcryptHash); err == nil {
		return fmt.Errorf("api key already used")
	}

	// Compare client's bcrypt hash against server's TOKEN
	if err := bcrypt.CompareHashAndPassword([]byte(clientBcryptHash), []byte(config.Use.App.Key)); err != nil {
		return err
	}

	// Store in Redis to prevent reuse (10 minute window)
	return pkg.RedisAuth.Set(ctx, clientBcryptHash, []byte("used"), 10*time.Minute)
}
