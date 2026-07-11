package ratelimit

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type rule struct {
	Name      string
	Limit     int
	WindowSec int
}

type config struct {
	Enabled         bool
	DefaultLimit    int
	DefaultWindow   int
	LoginLimit      int
	LoginWindow     int
	CaptchaLimit    int
	CaptchaWindow   int
	SensitiveLimit  int
	SensitiveWindow int
}

func Middleware(serviceName string, rdb redis.UniversalClient, next http.Handler) http.Handler {
	cfg := loadConfig(config{
		Enabled:         true,
		DefaultLimit:    600,
		DefaultWindow:   60,
		LoginLimit:      10,
		LoginWindow:     60,
		CaptchaLimit:    60,
		CaptchaWindow:   60,
		SensitiveLimit:  180,
		SensitiveWindow: 60,
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cfg.Enabled || rdb == nil || r.Method == http.MethodOptions || r.URL.Path == "/ping" {
			next.ServeHTTP(w, r)
			return
		}
		selected := pickRule(cfg, r.URL.Path)
		allowed, retryAfter := allow(r.Context(), rdb, serviceName, selected, clientIP(r), r.URL.Path)
		if !allowed {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"code":0,"msg":"请求过于频繁，请稍后再试"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func allow(ctx context.Context, rdb redis.UniversalClient, serviceName string, r rule, ip, path string) (bool, int) {
	if r.Limit <= 0 || r.WindowSec <= 0 {
		return true, 0
	}
	now := time.Now().Unix()
	window := now / int64(r.WindowSec)
	pathHash := sha1.Sum([]byte(path))
	key := fmt.Sprintf("security:rl:%s:%s:%s:%s:%d", serviceName, r.Name, sanitizeKey(ip), hex.EncodeToString(pathHash[:8]), window)
	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return true, 0
	}
	if count == 1 {
		_ = rdb.Expire(ctx, key, time.Duration(r.WindowSec+2)*time.Second).Err()
	}
	if count > int64(r.Limit) {
		retryAfter := int(int64(r.WindowSec) - (now % int64(r.WindowSec)))
		if retryAfter <= 0 {
			retryAfter = r.WindowSec
		}
		return false, retryAfter
	}
	return true, 0
}

func pickRule(cfg config, path string) rule {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, "/passport/login"):
		return rule{Name: "login", Limit: cfg.LoginLimit, WindowSec: cfg.LoginWindow}
	case strings.Contains(lower, "/passport/captcha"):
		return rule{Name: "captcha", Limit: cfg.CaptchaLimit, WindowSec: cfg.CaptchaWindow}
	case strings.Contains(lower, "/secret/") || strings.Contains(lower, "/message/") || strings.Contains(lower, "/connection/"):
		return rule{Name: "sensitive", Limit: cfg.SensitiveLimit, WindowSec: cfg.SensitiveWindow}
	default:
		return rule{Name: "default", Limit: cfg.DefaultLimit, WindowSec: cfg.DefaultWindow}
	}
}

func clientIP(r *http.Request) string {
	if ip := firstForwardedIP(r.Header.Get("X-Forwarded-For")); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); net.ParseIP(ip) != nil {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if net.ParseIP(r.RemoteAddr) != nil {
		return r.RemoteAddr
	}
	return "unknown"
}

func firstForwardedIP(value string) string {
	for _, part := range strings.Split(value, ",") {
		ip := strings.TrimSpace(part)
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}

func sanitizeKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(":", "_", "/", "_", "\\", "_", " ", "_")
	return replacer.Replace(value)
}

func loadConfig(cfg config) config {
	cfg.Enabled = envBool("SECURITY_RATE_LIMIT_ENABLED", cfg.Enabled)
	cfg.DefaultLimit = envInt("SECURITY_RATE_LIMIT_MAX_REQUESTS", cfg.DefaultLimit)
	cfg.DefaultWindow = envInt("SECURITY_RATE_LIMIT_WINDOW_SEC", cfg.DefaultWindow)
	cfg.LoginLimit = envInt("SECURITY_RATE_LIMIT_LOGIN_MAX_REQUESTS", cfg.LoginLimit)
	cfg.LoginWindow = envInt("SECURITY_RATE_LIMIT_LOGIN_WINDOW_SEC", cfg.LoginWindow)
	cfg.CaptchaLimit = envInt("SECURITY_RATE_LIMIT_CAPTCHA_MAX_REQUESTS", cfg.CaptchaLimit)
	cfg.CaptchaWindow = envInt("SECURITY_RATE_LIMIT_CAPTCHA_WINDOW_SEC", cfg.CaptchaWindow)
	cfg.SensitiveLimit = envInt("SECURITY_RATE_LIMIT_SENSITIVE_MAX_REQUESTS", cfg.SensitiveLimit)
	cfg.SensitiveWindow = envInt("SECURITY_RATE_LIMIT_SENSITIVE_WINDOW_SEC", cfg.SensitiveWindow)
	return cfg
}

func envInt(key string, def int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func envBool(key string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}
