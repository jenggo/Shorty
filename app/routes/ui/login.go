package ui

import (
	"fmt"
	"shorty/types"
	"shorty/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

func OauthLogin(ctx fiber.Ctx) error {
	state := utils.GenerateState()

	sess, err := sessionStore.Get(ctx)
	if err != nil {
		log.Error().Caller().Err(err).Msg("failed to get session")
		return fmt.Errorf("failed to get session")
	}
	defer sess.Release()

	if !sess.Fresh() {
		if err := sess.Regenerate(); err != nil {
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
	}

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
