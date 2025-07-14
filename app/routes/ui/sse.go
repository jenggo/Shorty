package ui

import (
	"bufio"
	"context"
	"fmt"
	"shorty/pkg"
	"shorty/types"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

func SSE(ctx fiber.Ctx) error {
	sessionID, err := validateSession(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	log.Debug().Str("sessionID", *sessionID).Msg("connected SSE client")

	// Set headers
	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")

	done := make(chan bool)
	return ctx.SendStreamWriter(func(w *bufio.Writer) {
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
					if err.Error() != "connection closed" {
						log.Error().Caller().Err(err).Msg("failed to write data")
					}

					return
				}

				// usually because connection is closed, just return instead of showing log
				if err := w.Flush(); err != nil {
					return
				}
			case <-done:
				log.Debug().Str("sessionID", *sessionID).Msg("client disconnected")
				return
			}
		}
	})
}
