package app

import (
	"crypto/sha512"
	"fmt"
	"shorty/config"
	"shorty/pkg"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/keyauth/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func verifyKey() func(*fiber.Ctx) error {
	return keyauth.New(keyauth.Config{
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			if err := verifyAPIKey(key); err != nil {
				log.Error().Caller().Err(err).Str("path", c.Path()).Str("hash", key).Send()
				return false, keyauth.ErrMissingOrMalformedAPIKey
			}

			return true, nil
		},
		ErrorHandler: errHandler,
	})
}

func verifyAPIKey(hashed string) error {
	if _, err := pkg.RedisAuth.Get(hashed); err == nil {
		return fmt.Errorf("%s already used", hashed)
	}

	h := []byte(hashed)
	i := []byte(config.Use.App.Key)
	k := sha512.New()
	k.Write(i)
	hash := k.Sum(nil)

	if err := bcrypt.CompareHashAndPassword(h, hash); err != nil {
		return err
	}

	return pkg.RedisAuth.Set(hashed, hash, 10*time.Minute)
}
