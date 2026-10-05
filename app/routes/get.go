package routes

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"shorty/config"
	"shorty/pkg"

	"github.com/gofiber/fiber/v3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

func Get(ctx fiber.Ctx) error {
	shorturl := ctx.Params("shorty")

	// Get the short URL data
	realurl, err := pkg.Redis.Get(ctx.Context(), shorturl)
	if err != nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	// A link pointing at a GitHub repo's releases resolves to the download URL
	// of its latest linux/x86_64 asset instead of redirecting to the repo.
	if pkg.IsReleaseURL(realurl) {
		return redirectLatestRelease(ctx, shorturl, realurl)
	}

	// Check if this is an S3 URL with credentials
	s3Creds, err := pkg.Redis.GetS3Credentials(ctx.Context(), shorturl)
	if err == nil && s3Creds.Access != "" && s3Creds.Secret != "" {
		// This URL has S3 credentials - generate a presigned URL
		parsedURL, err := url.Parse(realurl)
		if err != nil {
			log.Error().Err(err).Str("url", realurl).Msg("failed to parse URL for presigning")
			return ctx.Redirect().Status(fiber.StatusPermanentRedirect).To(realurl)
		}

		// Extract relevant parts from the URL
		endpoint := parsedURL.Host
		pathParts := strings.SplitN(strings.TrimPrefix(parsedURL.Path, "/"), "/", 2)
		if len(pathParts) != 2 {
			log.Error().Str("path", parsedURL.Path).Msg("invalid S3 URL path format")
			return ctx.Redirect().Status(fiber.StatusPermanentRedirect).To(realurl)
		}

		bucket := pathParts[0]
		objectName := pathParts[1]

		// Initialize MinIO client with user credentials
		s3Client, err := minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(s3Creds.Access, s3Creds.Secret, ""),
			Secure: parsedURL.Scheme == "https",
		})

		if err != nil {
			log.Error().Err(err).Msg("failed to initialize S3 client")
			return ctx.Redirect().Status(fiber.StatusPermanentRedirect).To(realurl)
		}

		// Set expiry time (use app config or a default)
		expiry := 7 * 24 * time.Hour // 7 days default
		if config.Use.S3.Expired > 0 {
			expiry = config.Use.S3.Expired
		}

		// Generate presigned URL
		reqParams := make(url.Values)
		presignedURL, err := s3Client.PresignedGetObject(ctx.Context(), bucket, objectName, expiry, reqParams)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate presigned URL")
			return ctx.Redirect().Status(fiber.StatusPermanentRedirect).To(realurl)
		}

		// Redirect to the presigned URL
		return ctx.Redirect().Status(fiber.StatusPermanentRedirect).To(presignedURL.String())
	}

	// No S3 credentials, just redirect to the stored URL
	return ctx.Redirect().Status(fiber.StatusPermanentRedirect).To(realurl)
}

// redirectLatestRelease sends the visitor to the download URL of the latest
// release's linux/x86_64 asset, mapping resolution failures onto statuses.
func redirectLatestRelease(ctx fiber.Ctx, shorty, rawURL string) error {
	asset, err := pkg.LatestReleaseAsset(ctx.Context(), rawURL)
	if err != nil {
		log.Error().Err(err).Str("shorty", shorty).Str("url", rawURL).Send()

		status := fiber.StatusBadGateway
		switch {
		case errors.Is(err, pkg.ErrNoRelease), errors.Is(err, pkg.ErrNoAsset):
			status = fiber.StatusNotFound
		case errors.Is(err, pkg.ErrRateLimited):
			status = fiber.StatusServiceUnavailable
		}

		return ctx.SendStatus(status)
	}

	return ctx.Redirect().Status(fiber.StatusFound).To(asset)
}
