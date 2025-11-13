package ui

import (
	"shorty/types"

	"github.com/gofiber/fiber/v3"
)

func GetAuthConfig(ctx fiber.Ctx) error {
	authMethods := GetAuthMethods()

	return ctx.JSON(types.Response{
		Error: false,
		Data: fiber.Map{
			"oauth":    authMethods["oauth"],
			"userpass": authMethods["userpass"],
		},
	})
}
