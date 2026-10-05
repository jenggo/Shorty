package routes

import (
	"fmt"

	"shorty/config"
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
	// fasthttp's default 4KB read buffer is smaller than the response headers
	// some hosts return (GitHub's CSP headers alone exceed it), which surfaces
	// as a read error on an otherwise reachable URL.
	cc.FasthttpClient().ReadBufferSize = 64 * 1024
	testUrl, err := cc.Head(body.Url)
	if err != nil {
		return fmt.Errorf("error when reach %s: %w", body.Url, err)
	}

	statusCode := testUrl.StatusCode()
	if statusCode == 404 || statusCode >= 500 {
		return fmt.Errorf("cannot reach %s, status code: %d", body.Url, statusCode)
	}

	if body.Shorty == "" {
		body.Shorty = utils.HumanFriendlyEnglishString(8)
	}

	// Permanent links are opt-in. When they are disallowed, a request without an
	// expiry falls back to the default instead of being rejected, so clients that
	// never send one keep working.
	if body.Expired <= 0 && !config.Use.App.AllowPermanent {
		body.Expired = config.Use.App.DefaultExpired
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
