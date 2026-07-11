package security

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
	imredis "pte_live_im/pkg/redis"
)

type rule struct {
	Name      string
	Limit     int
	WindowSec int
}

type config struct {
	Enabled       bool
	APILimit      int
	APIWindow     int
	WSLimit       int
	WSWindow      int
	MessagePerSec int
}

func HTTPMiddleware(kind string, next http.HandlerFunc) http.HandlerFunc {
	cfg := loadConfig(defaultConfig())
	return func(w http.ResponseWriter, r *http.Request) {
		if !cfg.Enabled || r.Method == http.MethodOptions || r.URL.Path == "/ping" {
			next(w, r)
			return
		}
		rule := pickRule(cfg, kind)
		allowed, retryAfter := allow(r.Context(), imredis.Client(), "im", rule, clientIP(r), r.URL.Path)
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "请求过于频繁，请稍后再试", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func MessagePerSecondLimit() int {
	cfg := loadConfig(defaultConfig())
	if !cfg.Enabled {
		return 0
	}
	return cfg.MessagePerSec
}

func allow(ctx context.Context, rdb redis.UniversalClient, serviceName string, r rule, ip, path string) (bool, int) {
	if rdb == nil || r.Limit <= 0 || r.WindowSec <= 0 {
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

func pickRule(cfg config, kind string) rule {
	if strings.EqualFold(kind, "ws") {
		return rule{Name: "ws", Limit: cfg.WSLimit, WindowSec: cfg.WSWindow}
	}
	return rule{Name: "api", Limit: cfg.APILimit, WindowSec: cfg.APIWindow}
}

func defaultConfig() config {
	return config{
		Enabled:       true,
		APILimit:      1200,
		APIWindow:     60,
		WSLimit:       60,
		WSWindow:      60,
		MessagePerSec: 30,
	}
}

func loadConfig(cfg config) config {
	cfg.Enabled = envBool("SECURITY_RATE_LIMIT_ENABLED", cfg.Enabled)
	cfg.APILimit = envInt("IM_RATE_LIMIT_API_MAX_REQUESTS", envInt("SECURITY_RATE_LIMIT_MAX_REQUESTS", cfg.APILimit))
	cfg.APIWindow = envInt("IM_RATE_LIMIT_API_WINDOW_SEC", envInt("SECURITY_RATE_LIMIT_WINDOW_SEC", cfg.APIWindow))
	cfg.WSLimit = envInt("IM_RATE_LIMIT_WS_MAX_REQUESTS", cfg.WSLimit)
	cfg.WSWindow = envInt("IM_RATE_LIMIT_WS_WINDOW_SEC", cfg.WSWindow)
	cfg.MessagePerSec = envInt("IM_WS_MESSAGE_MAX_PER_SEC", cfg.MessagePerSec)
	return cfg
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
