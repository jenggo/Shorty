package types

import "time"

type Response struct {
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
}

type Shorten struct {
	Url     string        `json:"url"`
	Shorty  string        `json:"shorty,omitempty"`
	Expired time.Duration `json:"expired,omitempty"`
}
