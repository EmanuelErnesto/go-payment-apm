package main

import (
	"errors"
	"fmt"
	"net/url"

	"time"

	"github.com/vrischmann/envconfig"
	"go.uber.org/zap/zapcore"
)

type config struct {
	HTTP     httpConfig
	Database databaseConfig
	Log      logConfig
	APM      apmConfig
}

type httpConfig struct {
	Port           string        `envconfig:"HTTP_PORT,default=8080"`
	Mode           string        `envconfig:"GIN_MODE,default=release"`
	SwaggerEnabled bool          `envconfig:"SWAGGER_ENABLED,default=true"`
	SimulatedDelay time.Duration `envconfig:"PAYMENT_SIMULATED_DELAY,default=2s"`
	WriteTimeout   time.Duration `envconfig:"HTTP_WRITE_TIMEOUT,default=30s"`
}

type databaseConfig struct {
	Host     string `envconfig:"DB_HOST,default=localhost"`
	Port     string `envconfig:"DB_PORT,default=5432"`
	User     string `envconfig:"DB_USER,default=payments_app"`
	Password string `envconfig:"DB_PASSWORD"`
	Name     string `envconfig:"DB_NAME,default=payments"`
	SSLMode  string `envconfig:"DB_SSLMODE,default=disable"`
}

type logConfig struct {
	Level logLevel `envconfig:"LOG_LEVEL,default=info"`
}

type apmConfig struct {
	ServiceName string      `envconfig:"ELASTIC_APM_SERVICE_NAME,default=payments-api"`
	Environment string      `envconfig:"ELASTIC_APM_ENVIRONMENT,default=local"`
	ServerURL   absoluteURL `envconfig:"ELASTIC_APM_SERVER_URL,default=http://localhost:8200"`
	SecretToken string      `envconfig:"ELASTIC_APM_SECRET_TOKEN,optional"`
}

func loadConfig() (config, error) {
	var cfg config
	if err := envconfig.Init(&cfg); err != nil {
		return config{}, err
	}
	return cfg, cfg.validate()
}

// validate holds only cross-field rules; per-variable parsing and presence
// checks are already enforced by envconfig.
func (c config) validate() error {
	// The delayed endpoint would always be cut by the server before answering.
	if c.HTTP.SimulatedDelay >= c.HTTP.WriteTimeout {
		return errors.New("PAYMENT_SIMULATED_DELAY must be lower than HTTP_WRITE_TIMEOUT")
	}
	return nil
}

func (d databaseConfig) DSN() string {
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(d.User, d.Password),
		Host:     d.Host + ":" + d.Port,
		Path:     d.Name,
		RawQuery: url.Values{"sslmode": {d.SSLMode}}.Encode(),
	}
	return dsn.String()
}

// logLevel adapts zapcore.Level to envconfig.Unmarshaler.
type logLevel zapcore.Level

func (l *logLevel) Unmarshal(s string) error {
	level, err := zapcore.ParseLevel(s)
	if err != nil {
		return err
	}
	*l = logLevel(level)
	return nil
}

func (l logLevel) Zap() zapcore.Level { return zapcore.Level(l) }

// absoluteURL implements envconfig.Unmarshaler and rejects relative URLs, which
// url.Parse would otherwise accept silently (e.g. "apm-server:8200").
type absoluteURL struct{ *url.URL }

func (u *absoluteURL) Unmarshal(s string) error {
	parsed, err := url.Parse(s)
	if err != nil {
		return err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("%q is not an absolute URL", s)
	}
	u.URL = parsed
	return nil
}
