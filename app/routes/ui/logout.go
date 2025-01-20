package ui

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

func Logout(ctx fiber.Ctx) error {
	sess, err := sessionStore.Get(ctx)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to get session")
		return fmt.Errorf("failed to get session")
	}
	defer sess.Release()

	if err := sess.Destroy(); err != nil {
		log.Error().Err(err).Msg("failed to destroy session")
		return fmt.Errorf("failed to logout")
	}

	return ctx.Render("logout", nil)
}
