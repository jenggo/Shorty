package pkg

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"shorty/utils"

	"github.com/gofiber/fiber/v3/client"
	"github.com/rs/zerolog/log"
)

const (
	ghReleaseCachePrefix = "ghrel:"

	// ghReleaseRevalidate is how long a resolved asset is trusted. Each
	// revalidation costs one request against GitHub's unauthenticated limit of
	// 60 per hour per IP, and conditional requests are counted too, so the
	// interval is the only thing that bounds the cost.
	ghReleaseRevalidate = time.Hour

	// ghReleaseCacheTTL outlives the revalidation interval so a stale entry is
	// still available to serve when GitHub cannot be reached.
	ghReleaseCacheTTL = 24 * time.Hour

	ghReleaseTimeout = 10 * time.Second
)

// Errors returned by LatestReleaseAsset so callers can pick a status code.
var (
	ErrNoRelease   = errors.New("no published release")
	ErrNoAsset     = errors.New("no linux/x86_64 asset in latest release")
	ErrRateLimited = errors.New("github api rate limited")
)

type ghReleaseCache struct {
	URL     string `json:"url"`
	Checked int64  `json:"checked"`
}

type ghRelease struct {
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// IsReleaseURL reports whether rawURL points at a GitHub repository's releases.
// Both https://github.com/owner/repo/releases and the canonical
// https://github.com/owner/repo/releases/latest form are accepted.
func IsReleaseURL(rawURL string) bool {
	_, ok := releaseRepo(rawURL)

	return ok
}

// releaseRepo extracts "owner/repo" from a GitHub releases URL.
func releaseRepo(rawURL string) (string, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}

	if u.Host != "github.com" && u.Host != "www.github.com" {
		return "", false
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 4 || parts[2] != "releases" {
		return "", false
	}

	if parts[3] != "latest" && parts[3] != "" {
		return "", false
	}

	if parts[0] == "" || parts[1] == "" {
		return "", false
	}

	return parts[0] + "/" + parts[1], true
}

// LatestReleaseAsset returns the download URL of the linux/x86_64 asset of the
// repository's latest release. The result is cached per repository, so several
// short links to the same repo share one entry and cost one API request.
func LatestReleaseAsset(ctx context.Context, rawURL string) (string, error) {
	repo, ok := releaseRepo(rawURL)
	if !ok {
		return "", fmt.Errorf("%s is not a github releases url", rawURL)
	}

	cache := readReleaseCache(repo)
	if cache != nil && time.Since(time.Unix(cache.Checked, 0)) < ghReleaseRevalidate {
		return cache.URL, nil
	}

	asset, err := fetchLatestAsset(ctx, repo)
	if err != nil {
		// A stale entry beats failing, but only when GitHub is rate limiting us:
		// a missing release or a missing asset is a definitive answer.
		if cache != nil && errors.Is(err, ErrRateLimited) {
			log.Warn().Err(err).Str("repo", repo).Msg("serving stale github release asset")

			return cache.URL, nil
		}

		return "", err
	}

	writeReleaseCache(repo, ghReleaseCache{URL: asset, Checked: time.Now().Unix()})

	return asset, nil
}

func fetchLatestAsset(ctx context.Context, repo string) (string, error) {
	cc := client.New()
	cc.SetTimeout(ghReleaseTimeout)

	resp, err := cc.Get("https://api.github.com/repos/"+repo+"/releases/latest", client.Config{
		Ctx: ctx,
		Header: map[string]string{
			"Accept":     "application/vnd.github+json",
			"User-Agent": "shorty",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to reach github api: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusNotFound:
		return "", ErrNoRelease
	case http.StatusForbidden, http.StatusTooManyRequests:
		return "", ErrRateLimited
	case http.StatusOK:
	default:
		return "", fmt.Errorf("github api returned status %d", resp.StatusCode())
	}

	var release ghRelease
	if err := utils.FromJSON(resp.Body(), &release); err != nil {
		return "", fmt.Errorf("failed to decode github release: %w", err)
	}

	for _, a := range release.Assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, "linux") && strings.Contains(name, "x86_64") {
			return a.URL, nil
		}
	}

	return "", ErrNoAsset
}

func readReleaseCache(repo string) *ghReleaseCache {
	data, err := Redis.Get(context.Background(), ghReleaseCachePrefix+repo)
	if err != nil {
		return nil
	}

	var cache ghReleaseCache
	if err := utils.FromJSON([]byte(data), &cache); err != nil {
		log.Warn().Err(err).Str("repo", repo).Msg("discarding malformed github release cache entry")

		return nil
	}

	return &cache
}

func writeReleaseCache(repo string, cache ghReleaseCache) {
	if err := Redis.Set(context.Background(), ghReleaseCachePrefix+repo, string(utils.ToJSON(cache)), ghReleaseCacheTTL); err != nil {
		log.Warn().Err(err).Str("repo", repo).Msg("failed to cache github release asset")
	}
}
