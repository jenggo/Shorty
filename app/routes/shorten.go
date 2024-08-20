package routes

import (
	"fmt"

	"shorty/pkg"
	"shorty/types"

	"github.com/gofiber/fiber/v2"
	"github.com/twharmon/gouid"
)

func Shorten(ctx *fiber.Ctx) error {
	var body types.Shorten

	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	if body.Url == "" {
		return fmt.Errorf("url cannot be empty")
	}

	// Check that url
	testUrl := fiber.Head(body.Url)
	statusCode, _, errs := testUrl.Bytes()
	if len(errs) > 0 && statusCode >= 500 {
		return fmt.Errorf("cannot reach %s, status code: %d", body.Url, statusCode)
	}

	if body.Shorty == "" {
		shortUrl, err := pkg.Redis.Get(ctx.Context(), body.Url)
		if err == nil {
			return fmt.Errorf("%s already had shorten url %s", body.Url, shortUrl)
		}

		body.Shorty = gouid.String(8, gouid.Secure32Char)
	}

	if err := pkg.Redis.Set(ctx.Context(), body.Shorty, body.Url, body.Expired); err != nil {
		return err
	}

	shorty := fmt.Sprintf("%s/%s", ctx.BaseURL(), body.Shorty)

	return ctx.JSON(types.Response{
		Error:   false,
		Message: shorty,
	})
}
