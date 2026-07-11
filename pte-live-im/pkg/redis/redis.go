package redis

import (
	"context"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"pte_live_im/pkg/setting"
)

var (
	client redis.UniversalClient
	once   sync.Once
)

func Client() redis.UniversalClient {
	once.Do(func() {
		cfg := setting.RedisSetting
		if len(cfg.Addrs) == 0 && strings.TrimSpace(cfg.Addr) == "" {
			return
		}
		addrs := cfg.Addrs
		if len(addrs) == 0 {
			addrs = []string{cfg.Addr}
		}
		if strings.EqualFold(cfg.Mode, "cluster") && len(addrs) > 1 {
			client = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    addrs,
				Password: cfg.Password,
			})
			return
		}
		addr := cfg.Addr
		if addr == "" {
			addr = addrs[0]
		}
		client = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	})
	return client
}

func Enabled() bool {
	return (setting.RedisSetting.Addr != "" || len(setting.RedisSetting.Addrs) > 0) && Client() != nil
}

func Ping(ctx context.Context) error {
	if !Enabled() {
		return nil
	}
	return Client().Ping(ctx).Err()
}
