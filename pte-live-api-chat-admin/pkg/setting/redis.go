package setting

import (
	"os"
	"strconv"
	"strings"
)

func normalizeRedis() {
	applyRedisEnv()
	if len(Redis.Addrs) == 0 && strings.TrimSpace(Redis.Addr) != "" {
		Redis.Addrs = []string{strings.TrimSpace(Redis.Addr)}
	}
	addrs := make([]string, 0, len(Redis.Addrs))
	for _, addr := range Redis.Addrs {
		if v := strings.TrimSpace(addr); v != "" {
			addrs = append(addrs, v)
		}
	}
	Redis.Addrs = addrs
	if len(Redis.Addrs) > 0 && strings.TrimSpace(Redis.Addr) == "" {
		Redis.Addr = Redis.Addrs[0]
	}
	mode := strings.ToLower(strings.TrimSpace(Redis.Mode))
	if mode == "" {
		if len(Redis.Addrs) > 1 {
			mode = "cluster"
		} else {
			mode = "single"
		}
	}
	Redis.Mode = mode
}

func applyRedisEnv() {
	if v := strings.TrimSpace(os.Getenv("REDIS_ADDR")); v != "" {
		Redis.Addr = v
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_PASSWORD")); v != "" {
		Redis.Password = v
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_MODE")); v != "" {
		Redis.Mode = v
	}
	if v := strings.TrimSpace(os.Getenv("REDIS_DB")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			Redis.DB = n
		}
	}
	addrs := make([]string, 0, 8)
	if v := strings.TrimSpace(os.Getenv("REDIS_ADDR")); v != "" {
		addrs = append(addrs, v)
	}
	for i := 2; i <= 8; i++ {
		if v := strings.TrimSpace(os.Getenv("REDIS_ADDR_" + strconv.Itoa(i))); v != "" {
			addrs = append(addrs, v)
		}
	}
	if len(addrs) > 0 {
		Redis.Addrs = addrs
	}
}
