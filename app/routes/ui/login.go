package ui

import (
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

	sess.Set("oauth_state", state)
	if err := sess.Save(); err != nil {
		log.Error().Err(err).Msg("failed to save session")
		return ctx.Status(fiber.StatusInternalServerError).JSON(types.Response{
			Error:   true,
			Message: "Failed to save login session",
		})
	}

	url := oauthConfig.AuthCodeURL(state)
	return ctx.JSON(types.Response{
		Error:   false,
		Message: url,
	})
}
