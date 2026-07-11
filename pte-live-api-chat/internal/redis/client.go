package redis

import (
	"context"
	"strings"

	goredis "github.com/redis/go-redis/v9"
	"pte_live_api_chat/pkg/setting"
)

func NewClient() goredis.UniversalClient {
	cfg := setting.Redis
	if len(cfg.Addrs) == 0 && strings.TrimSpace(cfg.Addr) == "" {
		return nil
	}
	addrs := cfg.Addrs
	if len(addrs) == 0 {
		addrs = []string{cfg.Addr}
	}
	if strings.EqualFold(cfg.Mode, "cluster") && len(addrs) > 1 {
		return goredis.NewClusterClient(&goredis.ClusterOptions{
			Addrs:    addrs,
			Password: cfg.Password,
		})
	}
	addr := cfg.Addr
	if addr == "" {
		addr = addrs[0]
	}
	return goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

func Ping(ctx context.Context, client goredis.UniversalClient) error {
	if client == nil {
		return nil
	}
	return client.Ping(ctx).Err()
}
