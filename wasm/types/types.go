package types

import (
	"net/http"
	"net/http/cookiejar"
	"time"
)

const API_BASE_URL = "https://u.nusatek.dev"

var jar, _ = cookiejar.New(nil)
var DefaultClient = &http.Client{Jar: jar}

type ShortyData struct {
	Shorty  string        `json:"shorty"`
	File    string        `json:"file"`
	URL     string        `json:"url"`
	Expired time.Duration `json:"expired"`
}

type APIResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}
