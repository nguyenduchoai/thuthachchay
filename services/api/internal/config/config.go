package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv         string `mapstructure:"APP_ENV"`
	LogLevel       string `mapstructure:"LOG_LEVEL"`
	HTTPListenAddr string `mapstructure:"HTTP_LISTEN_ADDR"`

	DatabaseURL             string        `mapstructure:"DATABASE_URL"`
	DatabaseMaxOpenConns    int           `mapstructure:"DATABASE_MAX_OPEN_CONNS"`
	DatabaseMaxIdleConns    int           `mapstructure:"DATABASE_MAX_IDLE_CONNS"`
	DatabaseConnMaxLifetime time.Duration `mapstructure:"DATABASE_CONN_MAX_LIFETIME"`

	RedisURL string `mapstructure:"REDIS_URL"`

	JWTPrivateKeyPath string        `mapstructure:"JWT_PRIVATE_KEY_PATH"`
	JWTPublicKeyPath  string        `mapstructure:"JWT_PUBLIC_KEY_PATH"`
	JWTAccessTTL      time.Duration `mapstructure:"JWT_ACCESS_TTL"`
	JWTRefreshTTL     time.Duration `mapstructure:"JWT_REFRESH_TTL"`

	ZaloAppID          string `mapstructure:"ZALO_APP_ID"`
	ZaloAppSecret      string `mapstructure:"ZALO_APP_SECRET"`
	ZaloOAID           string `mapstructure:"ZALO_OA_ID"`
	ZaloOAuthRedirect  string `mapstructure:"ZALO_OAUTH_REDIRECT"`
	StravaClientID     string `mapstructure:"STRAVA_CLIENT_ID"`
	StravaClientSecret string `mapstructure:"STRAVA_CLIENT_SECRET"`
	StravaRedirectURI  string `mapstructure:"STRAVA_REDIRECT_URI"`
	StravaVerifyToken  string `mapstructure:"STRAVA_VERIFY_TOKEN"`

	AntiFraudMaxDailySteps      int `mapstructure:"ANTIFRAUD_MAX_DAILY_STEPS"`
	AntiFraudMinCadenceMs       int `mapstructure:"ANTIFRAUD_MIN_CADENCE_PERIOD_MS"`
	AntiFraudMaxCadenceMs       int `mapstructure:"ANTIFRAUD_MAX_CADENCE_PERIOD_MS"`
	AntiFraudAutoSuspendScore   int `mapstructure:"ANTIFRAUD_AUTO_SUSPEND_SCORE"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	v.SetDefault("APP_ENV", "dev")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("HTTP_LISTEN_ADDR", ":8080")
	v.SetDefault("DATABASE_MAX_OPEN_CONNS", 25)
	v.SetDefault("DATABASE_MAX_IDLE_CONNS", 5)
	v.SetDefault("DATABASE_CONN_MAX_LIFETIME", "5m")
	v.SetDefault("JWT_ACCESS_TTL", "15m")
	v.SetDefault("JWT_REFRESH_TTL", "720h")
	v.SetDefault("ANTIFRAUD_MAX_DAILY_STEPS", 60000)
	v.SetDefault("ANTIFRAUD_MIN_CADENCE_PERIOD_MS", 300)
	v.SetDefault("ANTIFRAUD_MAX_CADENCE_PERIOD_MS", 900)
	v.SetDefault("ANTIFRAUD_AUTO_SUSPEND_SCORE", 90)

	// Bind các key cần đọc env (viper.AutomaticEnv không tự bind cho mapstructure
	// nếu chưa thấy key, nên gọi BindEnv tường minh).
	keys := []string{
		"APP_ENV", "LOG_LEVEL", "HTTP_LISTEN_ADDR",
		"DATABASE_URL", "DATABASE_MAX_OPEN_CONNS", "DATABASE_MAX_IDLE_CONNS", "DATABASE_CONN_MAX_LIFETIME",
		"REDIS_URL",
		"JWT_PRIVATE_KEY_PATH", "JWT_PUBLIC_KEY_PATH", "JWT_ACCESS_TTL", "JWT_REFRESH_TTL",
		"ZALO_APP_ID", "ZALO_APP_SECRET", "ZALO_OA_ID", "ZALO_OAUTH_REDIRECT",
		"STRAVA_CLIENT_ID", "STRAVA_CLIENT_SECRET", "STRAVA_REDIRECT_URI", "STRAVA_VERIFY_TOKEN",
		"ANTIFRAUD_MAX_DAILY_STEPS", "ANTIFRAUD_MIN_CADENCE_PERIOD_MS",
		"ANTIFRAUD_MAX_CADENCE_PERIOD_MS", "ANTIFRAUD_AUTO_SUSPEND_SCORE",
	}
	for _, k := range keys {
		_ = v.BindEnv(k)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validate(c *Config) error {
	if c.AppEnv == "" {
		return fmt.Errorf("APP_ENV is required")
	}
	switch c.AppEnv {
	case "dev", "staging", "prod":
	default:
		return fmt.Errorf("APP_ENV must be dev|staging|prod, got %q", c.AppEnv)
	}
	if c.AppEnv != "dev" {
		if c.DatabaseURL == "" {
			return fmt.Errorf("DATABASE_URL required outside dev")
		}
		if c.JWTPrivateKeyPath == "" || c.JWTPublicKeyPath == "" {
			return fmt.Errorf("JWT key paths required outside dev")
		}
	}
	return nil
}
