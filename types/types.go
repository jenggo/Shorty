package types

import "time"

type Response struct {
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
}

type Shorten struct {
	Url     string        `json:"url"`
	File    string        `json:"file,omitempty"`
	Shorty  string        `json:"shorty,omitempty"`
	Expired time.Duration `json:"expired,omitempty"`
}
