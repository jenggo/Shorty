package pkg

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3/client"
	"github.com/rs/zerolog/log"
)

const (
	// permanentLinkGuardInterval is how often permanent links are rechecked.
	permanentLinkGuardInterval = time.Hour
	linkCheckTimeout           = 10 * time.Second
)

// StartPermanentLinkGuard periodically removes short links that have no TTL and
// whose target has stopped working. Links with an expiry are left alone: Redis
// drops them on its own, and rechecking them would be wasted work.
func (r *redis) StartPermanentLinkGuard() {
	ticker := time.NewTicker(permanentLinkGuardInterval)
	go func() {
		defer ticker.Stop()

		r.guardPermanentLinks()
		for range ticker.C {
			r.guardPermanentLinks()
		}
	}()
}

func (r *redis) guardPermanentLinks() {
	ctx := context.Background()

	links, err := r.GetAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to list links for permanent-link guard")

		return
	}

	for _, link := range links {
		// GetAll reports Redis TTL as a duration, so a negative value means the
		// key has no expiry at all: only those links bypassed the TTL.
		if link.Expired >= 0 {
			continue
		}

		if !linkDead(link.Url) {
			continue
		}

		if err := r.Del(ctx, link.Shorty); err != nil {
			log.Error().Err(err).Str("shorty", link.Shorty).Msg("failed to remove dead permanent link")

			continue
		}

		log.Info().Str("shorty", link.Shorty).Str("url", link.Url).Msg("removed dead permanent link")
	}
}

// linkDead reports whether a link's target has stopped working. Presigned URLs
// are checked by computing their signature expiry, which costs no network call;
// everything else is probed with a HEAD request.
func linkDead(rawURL string) bool {
	if u, err := url.Parse(rawURL); err == nil {
		if expired, ok := presignExpired(u.Query()); ok {
			return expired
		}
	}

	cc := client.New()
	cc.SetTimeout(linkCheckTimeout)
	// Same reason as the reachability check in the shorten handler: fasthttp's
	// 4KB default read buffer is smaller than some hosts' response headers.
	cc.FasthttpClient().ReadBufferSize = 64 * 1024

	resp, err := cc.Head(rawURL)
	if err != nil {
		// A refused connection or an unknown host means the target is gone. Any
		// other failure (timeout, reset, transient DNS) says nothing about the
		// link, so it is left in place.
		if isTargetGone(err) {
			return true
		}

		log.Warn().Err(err).Str("url", rawURL).Msg("failed to check permanent link")

		return false
	}

	return resp.StatusCode() >= http.StatusBadRequest
}

// isTargetGone reports whether an error proves the host cannot serve the
// target, as opposed to a failure that may be temporary.
func isTargetGone(err error) bool {
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}

	dnsErr, ok := errors.AsType[*net.DNSError](err)
	if ok {
		return dnsErr.IsNotFound
	}

	return false
}

// presignExpired reads the SigV4 expiry from a presigned URL's query string.
// The second result reports whether the URL carries an expiry at all.
func presignExpired(q url.Values) (expired, ok bool) {
	date, expires := q.Get("X-Amz-Date"), q.Get("X-Amz-Expires")
	if date == "" || expires == "" {
		return false, false
	}

	start, err := time.Parse("20060102T150405Z", date)
	if err != nil {
		return false, false
	}

	seconds, err := strconv.Atoi(expires)
	if err != nil {
		return false, false
	}

	return time.Now().After(start.Add(time.Duration(seconds) * time.Second)), true
}
