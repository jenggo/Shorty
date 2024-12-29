package config

const (
	AppName    string = "Shorty"
	AppVersion string = "v0.0.4"
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
}
