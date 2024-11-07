package app

import (
	"context"
	"fmt"
	"time"

	"shorty/config"
	"shorty/pkg"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/keyauth/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/blake2b"
)

func verifyKey() func(*fiber.Ctx) error {
	return keyauth.New(keyauth.Config{
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			if err := verifyAPIKey(c.Context(), key); err != nil {
				log.Error().Caller().Err(err).Str("path", c.Path()).Str("hash", key).Send()
				return false, keyauth.ErrMissingOrMalformedAPIKey
			}

			return true, nil
		},
		ErrorHandler: errHandler,
	})
}

func verifyAPIKey(ctx context.Context, hashed string) error {
	if config.Use.App.Token != "" && hashed == config.Use.App.Token {
		return nil
	}

	if _, err := pkg.RedisAuth.Get(ctx, hashed); err == nil {
		return fmt.Errorf("%s already used", hashed)
	}

	h := []byte(hashed)
	i := []byte(config.Use.App.Key)
	k, err := blake2b.New256(nil)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return err
	}
	k.Write(i)
	hash := k.Sum(nil)

	if err := bcrypt.CompareHashAndPassword(h, hash); err != nil {
		return err
	}

	return pkg.RedisAuth.Set(ctx, hashed, hash, 10*time.Minute)
}
