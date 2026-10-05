package config

import "time"

const AppName string = "Shorty"

var Use config

type config struct {
	App struct {
		Listen     string `yaml:"listen" env:"LISTEN" default:":1106"`
		PPROF      string `yaml:"pprof" env:"PPROF"`
		LogLevel   int8   `yaml:"log_level" env:"LOG_LEVEL" default:"2"` // 0: debug, 1: info, 2: warning, 3: error, 4: fatal, 5: panic
		Cloudflare bool   `yaml:"cloudflare" env:"CLOUDFLARE" default:"true"`
		Key        string `yaml:"key" env:"KEY" validate:"required"`
		Token      string `yaml:"token" env:"TOKEN"`
		SentryURL  string `yaml:"sentry_url" env:"SENTRY_URL"`
		// SentryTracesSampleRate is the share of requests sent to Sentry as
		// transactions. Only used when SentryURL is set.
		SentryTracesSampleRate float64 `yaml:"sentry_traces_sample_rate" env:"SENTRY_TRACES_SAMPLE_RATE" default:"1.0"`
		Auth                   struct {
			User     string `yaml:"user" env:"AUTH_USER" default:"admin"`
			Password string `yaml:"password" env:"AUTH_PASSWORD" validate:"required"`
		} `yaml:"auth"`
		BaseURL string `yaml:"base_url" env:"BASE_URL" default:"https://u.nusatek.dev"`
		// AllowPermanent permits shorty links with no TTL. When false, a link
		// created without an expiry gets DefaultExpired instead.
		AllowPermanent bool          `yaml:"allow_permanent" env:"ALLOW_PERMANENT" default:"true"`
		DefaultExpired time.Duration `yaml:"default_expired" env:"DEFAULT_EXPIRED" default:"24h"`
	} `yaml:"app"`

	Redis struct {
		Host     string `yaml:"host" env:"REDIS_HOST" default:"127.0.0.1"`
		Port     string `yaml:"port" env:"REDIS_PORT" default:"6379"`
		Password string `yaml:"password" env:"REDIS_PASSWORD"`
		DB       struct {
			Main int `yaml:"main" env:"REDIS_DB" default:"0"`
			Auth int `yaml:"auth" env:"REDIS_DB_AUTH" default:"1"`
		} `yaml:"db"`
	} `yaml:"redis"`

	Oauth struct {
		Enable       bool   `yaml:"enable" env:"OAUTH_ENABLE" default:"false"`
		ClientID     string `yaml:"client_id" env:"OAUTH_CLIENT_ID"`
		ClientSecret string `yaml:"client_secret" env:"OAUTH_CLIENT_SECRET"`
		// RedirectURI  string `yaml:"redirect_uri" env:"OAUTH_REDIRECT_URI"`
		BaseURL string `yaml:"base_url" env:"OAUTH_BASE_URL"`
	} `yaml:"oauth"`

	S3 struct {
		Enable   bool   `yaml:"enable" env:"S3_ENABLE" default:"false"`
		Endpoint string `yaml:"endpoint" env:"S3_ENDPOINT"`
		Bucket   string `yaml:"bucket" env:"S3_BUCKET"`
		Key      struct {
			Access string `yaml:"access" env:"S3_ACCESS"`
			Secret string `yaml:"secret" env:"S3_SECRET"`
		} `yaml:"key"`
		Tracing         bool          `yaml:"tracing" env:"tracing" default:"false"`
		Expired         time.Duration `yaml:"expired" env:"S3_EXPIRED" default:"12h"`
		CleanupInterval time.Duration `yaml:"cleanup_interval" env:"S3_CLEANUP_INTERVAL" default:"1h"`
	} `yaml:"s3"`
}
