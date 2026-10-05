package app

import (
	"strings"

	"shorty/config"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v3"
)

// sentryTracing reports each request as a Sentry transaction. sentry-go ships
// no Fiber integration, so this mirrors what its net/http middleware does.
//
// Streaming endpoints are skipped: their handlers stay open for the life of the
// connection, so wrapping them would leave one transaction open for days.
func sentryTracing(ctx fiber.Ctx) error {
	if config.Use.App.SentryURL == "" {
		return ctx.Next()
	}

	if isStreamingRequest(ctx.Path()) {
		return ctx.Next()
	}

	reqCtx := ctx.Context()
	hub := sentry.GetHubFromContext(reqCtx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
		reqCtx = sentry.SetHubOnContext(reqCtx, hub)
	}

	transaction := sentry.StartTransaction(reqCtx, ctx.Route().Path,
		sentry.ContinueTrace(hub, ctx.Get(sentry.SentryTraceHeader), ctx.Get(sentry.SentryBaggageHeader)),
		sentry.WithOpName("http.server"),
		sentry.WithTransactionSource(sentry.SourceRoute),
		sentry.WithSpanOrigin(sentry.SpanOriginStdLib),
	)
	transaction.SetData("http.request.method", ctx.Method())

	ctx.SetContext(sentry.SetHubOnContext(transaction.Context(), hub))

	err := ctx.Next()

	// Routing has resolved by now, so the route pattern is available. Reading it
	// before Next yields "/" for every request.
	transaction.Name = ctx.Route().Path
	status := ctx.Response().StatusCode()
	transaction.Status = sentry.HTTPtoSpanStatus(status)
	transaction.SetData("http.response.status_code", status)
	transaction.Finish()

	return err
}

func isStreamingRequest(path string) bool {
	return path == "/events" ||
		strings.HasSuffix(path, "/sse") ||
		strings.HasSuffix(path, "/stream") ||
		strings.Contains(path, "/stream/")
}
