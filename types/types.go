package types

import "time"

type Response struct {
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type S3Credentials struct {
	Access string `json:"key_access,omitempty"`
	Secret string `json:"key_secret,omitempty"`
}

type Shorten struct {
	Url     string        `json:"url"`
	File    string        `json:"file,omitempty"`
	Shorty  string        `json:"shorty,omitempty"`
	Expired time.Duration `json:"expired,omitempty"`
	S3Key   S3Credentials `json:"s3_credentials,omitzero"`
}
