package ui

import (
	"errors"
	"shorty/config"
	"shorty/types"

	"github.com/gofiber/fiber/v3"
)

func validateSession(ctx fiber.Ctx, returnName ...bool) (*string, error) {
	sess, err := sessionStore.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer sess.Release()

	if _, ok := sess.Get("name").(string); !ok {
		return nil, errors.New("unauthorized access")
	}

	ret := sess.ID()
	if len(returnName) > 0 && returnName[0] {
		ret = sess.Get("name").(string)
	}

	return &ret, nil
}

func CheckSession(ctx fiber.Ctx) error {
	name, err := validateSession(ctx, true)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: "No valid session",
		})
	}

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
