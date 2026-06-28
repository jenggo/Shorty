package types

import "time"

type Response struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   bool   `json:"error"`
}

type S3Credentials struct {
	Access string `json:"key_access,omitempty"`
	Secret string `json:"key_secret,omitempty"`
}

type Shorten struct {
	S3Key   S3Credentials `json:"s3_credentials,omitzero"`
	Url     string        `json:"url"`
	File    string        `json:"file,omitempty"`
	Shorty  string        `json:"shorty,omitempty"`
	Expired time.Duration `json:"expired,omitempty"`
}
