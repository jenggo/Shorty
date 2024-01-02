package routes

import (
	"fmt"
	"shorty/pkg"
	"shorty/types"

	"github.com/gofiber/fiber/v2"
)

func Delete(ctx *fiber.Ctx) error {
	shorturl := ctx.Params("shorty")
	if _, err := pkg.Redis.Get(shorturl); err != nil {
		return err
	}

	if err := pkg.Redis.Del(shorturl); err != nil {
		return err
	}

	return ctx.JSON(types.Response{
		Error:   false,
		Message: fmt.Sprintf("%s deleted", shorturl),
	})
}
