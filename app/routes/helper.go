package routes

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog/log"
)

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
