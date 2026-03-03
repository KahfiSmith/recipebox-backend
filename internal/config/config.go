package config

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env                string
	HTTPAddr           string
	DatabaseURL        string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	BcryptCost         int
	GracefulShutdownMs int
	FrontendBaseURL    string
	AuthDebugExposeTokens bool
	TrustedProxyCIDRs  []*net.IPNet
	SMTPHost           string
	SMTPPort           int
	SMTPUsername       string
	SMTPPassword       string
	SMTPFromEmail      string
	SMTPFromName       string
	AuthRateLimitPerMinute int
}

func Load() (Config, error) {
	loadDotEnvIfPresent()

	cfg := Config{
		Env:                getEnv("APP_ENV", "development"),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:        strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisAddr:          strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		RedisDB:            getInt("REDIS_DB", 0),
		JWTSecret:          strings.TrimSpace(os.Getenv("JWT_SECRET")),
		AccessTokenTTL:     getDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:    getDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		BcryptCost:         getInt("BCRYPT_COST", 12),
		GracefulShutdownMs: getInt("GRACEFUL_SHUTDOWN_MS", 10000),
		FrontendBaseURL:    strings.TrimSpace(os.Getenv("FRONTEND_BASE_URL")),
		SMTPHost:           strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:           getInt("SMTP_PORT", 587),
		SMTPUsername:       strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:       strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		SMTPFromEmail:      strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")),
		SMTPFromName:       strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
		AuthRateLimitPerMinute: getInt("AUTH_RATE_LIMIT_PER_MINUTE", 30),
	}
	cfg.AuthDebugExposeTokens = getBool("AUTH_DEBUG_EXPOSE_TOKENS", cfg.Env != "production")

	trustedProxyCIDRs, err := parseTrustedProxyCIDRs(os.Getenv("TRUSTED_PROXY_CIDRS"))
	if err != nil {
		return Config{}, err
	}
	cfg.TrustedProxyCIDRs = trustedProxyCIDRs

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR is required")
	}
	if cfg.RedisDB < 0 {
		return Config{}, fmt.Errorf("REDIS_DB must be zero or greater")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.BcryptCost < 10 || cfg.BcryptCost > 14 {
		return Config{}, fmt.Errorf("BCRYPT_COST must be between 10 and 14")
	}
	if cfg.AccessTokenTTL <= 0 || cfg.RefreshTokenTTL <= 0 {
		return Config{}, fmt.Errorf("token TTL must be positive")
	}
	if cfg.RefreshTokenTTL <= cfg.AccessTokenTTL {
		return Config{}, fmt.Errorf("REFRESH_TOKEN_TTL must be greater than ACCESS_TOKEN_TTL")
	}
	if cfg.GracefulShutdownMs < 1000 {
		return Config{}, fmt.Errorf("GRACEFUL_SHUTDOWN_MS must be at least 1000")
	}
	if cfg.AuthRateLimitPerMinute <= 0 {
		return Config{}, fmt.Errorf("AUTH_RATE_LIMIT_PER_MINUTE must be positive")
	}
	if cfg.SMTPHost != "" {
		if cfg.SMTPPort <= 0 || cfg.SMTPPort > 65535 {
			return Config{}, fmt.Errorf("SMTP_PORT must be between 1 and 65535")
		}
		if cfg.SMTPFromEmail == "" {
			return Config{}, fmt.Errorf("SMTP_FROM_EMAIL is required when SMTP_HOST is set")
		}
	}
	if !cfg.AuthDebugExposeTokens && cfg.SMTPHost == "" {
		return Config{}, fmt.Errorf("SMTP_HOST is required when AUTH_DEBUG_EXPOSE_TOKENS=false")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return v
}

func getDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	v, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return v
}

func getBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func loadDotEnvIfPresent() {
	envFile := ".env"
	if _, err := os.Stat(envFile); err != nil {
		return
	}

	file, err := os.Open(filepath.Clean(envFile))
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}

		value = strings.Trim(value, `"'`)

		// Keep shell-provided env values as highest priority.
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		_ = os.Setenv(key, value)
	}
}

func parseTrustedProxyCIDRs(raw string) ([]*net.IPNet, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	networks := make([]*net.IPNet, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "/") {
			_, network, err := net.ParseCIDR(part)
			if err != nil {
				return nil, fmt.Errorf("invalid TRUSTED_PROXY_CIDRS entry %q", part)
			}
			networks = append(networks, network)
			continue
		}

		ip := net.ParseIP(part)
		if ip == nil {
			return nil, fmt.Errorf("invalid TRUSTED_PROXY_CIDRS entry %q", part)
		}

		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		networks = append(networks, &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(bits, bits),
		})
	}

	return networks, nil
}
