package routes

import (
	"bufio"
	"context"
	"fmt"
	"shorty/pkg"
	"shorty/types"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

func SSEHandler(c *fiber.Ctx) error {
	if err := validateSession(c); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	sess, err := getSession(c)
	if err != nil {
		return err
	}
	sessionID := sess.ID()
	log.Debug().Str("sessionID", sessionID).Msg("connected SSE client")

	// Set headers
	c.Context().SetContentType("text/event-stream")
	c.Context().Response.Header.Set(fiber.HeaderCacheControl, "no-cache")
	c.Context().Response.Header.Set(fiber.HeaderConnection, fiber.HeaderKeepAlive)
	c.Context().Response.Header.Set(fiber.HeaderTransferEncoding, "chunked")
	c.Context().Response.Header.Set(fiber.HeaderAccessControlAllowHeaders, fiber.HeaderCacheControl)
	c.Context().Response.Header.Set(fiber.HeaderAccessControlAllowCredentials, "true")

	done := make(chan bool)
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer close(done)

		// Send connected event
		if _, err := fmt.Fprintf(w, "event: connected\ndata: true\n\n"); err != nil {
			log.Error().Err(err).Msg("failed to send connected event")
			return
		}
		if err := w.Flush(); err != nil {
			return
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				lists, err := pkg.Redis.GetAll(context.Background())
				if err != nil {
					log.Error().Caller().Err(err).Msg("failed to get data")
					continue
				}

				jsonData, err := json.Marshal(lists)
				if err != nil {
					log.Error().Caller().Err(err).Msg("failed to marshal data")
					continue
				}

				// Send keepalive comment
				if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
					log.Error().Caller().Err(err).Msg("failed to send keepalive")
					return
				}

				if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
					log.Error().Caller().Err(err).Msg("failed to write data")
					return
				}

				// usually because connection is closed, just return instead of showing log
				if err := w.Flush(); err != nil {
					return
				}
			case <-done:
				log.Debug().Str("sessionID", sessionID).Msg("client disconnected")
				return
			}
		}
	}))

	return nil
}
