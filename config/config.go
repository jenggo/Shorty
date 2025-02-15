package config

import (
	"time"
)

const (
	AppName    string = "Shorty"
	AppVersion string = "v0.0.6"
)

var Use config

type config struct {
	App struct {
		Listen     string `yaml:"listen" env:"LISTEN" env-default:":1106"`
		PPROF      string `yaml:"pprof" env:"PPROF"`
		LogLevel   int8   `yaml:"log_level" env:"LOG_LEVEL" env-default:"2"` // 0: debug, 1: info, 2: warning, 3: error, 4: fatal, 5: panic
		Cloudflare bool   `yaml:"cloudflare" env:"CLOUDFLARE" env-default:"true"`
		Key        string `yaml:"key" env:"KEY" env-required:"true"`
		Token      string `yaml:"token" env:"TOKEN"`
		Sentry     string `yaml:"sentry" env:"SENTRY"`
		Auth       struct {
			User     string `yaml:"user" env:"AUTH_USER" env-default:"admin"`
			Password string `yaml:"password" env:"AUTH_PASSWORD" env-required:"true"`
		} `yaml:"auth"`
		BaseURL string `yaml:"base_url" env:"BASE_URL" env-default:"https://u.nusatek.dev"`
	} `yaml:"app"`

	Redis struct {
		Host     string `yaml:"host" env:"REDIS_HOST" env-default:"127.0.0.1"`
		Port     string `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
		Password string `yaml:"password" env:"REDIS_PASSWORD"`
		DB       struct {
			Main int `yaml:"main" env:"REDIS_DB" env-default:"0"`
			Auth int `yaml:"auth" env:"REDIS_DB_AUTH" env-default:"1"`
		} `yaml:"db"`
	} `yaml:"redis"`

	Oauth struct {
		ClientID     string `yaml:"client_id" env:"OAUTH_CLIENT_ID"`
		ClientSecret string `yaml:"client_secret" env:"OAUTH_CLIENT_SECRET"`
		RedirectURI  string `yaml:"redirect_uri" env:"OAUTH_REDIRECT_URI"`
		BaseURL      string `yaml:"base_url" env:"OAUTH_BASE_URL"`
	} `yaml:"oauth"`

	S3 struct {
		Enable   bool   `yaml:"enable" env:"S3_ENABLE" env-default:"false"`
		Endpoint string `yaml:"endpoint" env:"S3_ENDPOINT"`
		Bucket   string `yaml:"bucket" env:"S3_BUCKET"`
		Key      struct {
			Access string `yaml:"access" env:"S3_ACCESS"`
			Secret string `yaml:"secret" env:"S3_SECRET"`
		} `yaml:"key"`
		Tracing         bool          `yaml:"tracing" env:"tracing" env-default:"false"`
		Expired         time.Duration `yaml:"expired" env:"S3_EXPIRED" env-default:"12h"`
		CleanupInterval time.Duration `yaml:"cleanup_interval" env:"S3_CLEANUP_INTERVAL" env-default:"1h"`
	} `yaml:"s3"`
}
