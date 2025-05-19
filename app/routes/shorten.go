package routes

import (
	"fmt"

	"shorty/pkg"
	"shorty/types"
	"shorty/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
)

func Shorten(ctx fiber.Ctx) error {
	var body types.Shorten
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}

	if body.Url == "" {
		return fmt.Errorf("url cannot be empty")
	}

	// Check that url
	cc := client.New()
	testUrl, err := cc.Head(body.Url)
	if err != nil {
		return fmt.Errorf("error when reach %s: %v", body.Url, err)
	}

	statusCode := testUrl.StatusCode()
	if statusCode == 404 || statusCode >= 500 {
		return fmt.Errorf("cannot reach %s, status code: %d", body.Url, statusCode)
	}

	if body.Shorty == "" {
		body.Shorty = utils.HumanFriendlyEnglishString(8)
	}

	// Check if S3 credentials are provided
	if body.S3Key.Access != "" && body.S3Key.Secret != "" {
		// Store URL with S3 credentials
		if err := pkg.Redis.SetWithS3Credentials(
			ctx.Context(), 
			body.Shorty, 
			body.Url, 
			body.S3Key, 
			body.Expired, 
			true,
		); err != nil {
			return err
		}
	} else {
		// Regular URL without S3 credentials
		if err := pkg.Redis.Set(ctx.Context(), body.Shorty, body.Url, body.Expired, true); err != nil {
			return err
		}
	}

	return ctx.JSON(types.Response{
		Error:   false,
		Message: fmt.Sprintf("%s/%s", ctx.BaseURL(), body.Shorty),
	})
}
