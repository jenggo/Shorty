package ui

import (
	"bufio"
	"context"
	"fmt"
	"shorty/pkg"
	"shorty/types"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

func writeSSEEvent(w *bufio.Writer, event string, data string) error {
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	return w.Flush()
}

func writeSSEComment(w *bufio.Writer, comment string) error {
	if _, err := fmt.Fprintf(w, ": %s\n\n", comment); err != nil {
		return err
	}
	return w.Flush()
}

func sendSSEData(w *bufio.Writer) error {
	lists, err := pkg.Redis.GetAll(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get data: %w", err)
	}

	jsonData, err := json.Marshal(lists)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := writeSSEComment(w, "keepalive"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(jsonData)); err != nil {
		return err
	}

	return w.Flush()
}

func SSE(ctx fiber.Ctx) error {
	sessionID, err := validateSession(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(types.Response{
			Error:   true,
			Message: err.Error(),
		})
	}

	log.Debug().Str("sessionID", *sessionID).Msg("connected SSE client")

	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")

	done := make(chan bool)
	return ctx.SendStreamWriter(func(w *bufio.Writer) {
		defer close(done)

		if err := writeSSEEvent(w, "connected", "true"); err != nil {
			log.Error().Err(err).Msg("failed to send connected event")
			return
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := sendSSEData(w); err != nil {
					if !strings.Contains(err.Error(), "connection closed") {
						log.Error().Caller().Err(err).Msg("failed to send SSE data")
					}
					return
				}
			case <-done:
				log.Debug().Str("sessionID", *sessionID).Msg("client disconnected")
				return
			}
		}
	})
}
