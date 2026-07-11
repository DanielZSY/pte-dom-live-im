package queue

import (
	"strings"

	"pte_live_im/pkg/pulsar"
	"pte_live_im/pkg/redis"
	"pte_live_im/pkg/setting"
)

func normalizeBackend(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pulsar":
		return "pulsar"
	case "both", "dual", "dual-write", "dualwrite":
		return "both"
	case "redis":
		return "redis"
	default:
		return "pulsar"
	}
}

func normalizeConsumeFrom(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pulsar":
		return "pulsar"
	case "redis":
		return "redis"
	default:
		return "pulsar"
	}
}

// ProduceRedis 是否写入 Redis List。
func ProduceRedis() bool {
	switch normalizeBackend(setting.QueueSetting.Backend) {
	case "pulsar":
		return false
	default:
		return true
	}
}

// ProducePulsar 是否写入 Pulsar。
func ProducePulsar() bool {
	switch normalizeBackend(setting.QueueSetting.Backend) {
	case "redis":
		return false
	default:
		return setting.PulsarSetting.Enabled
	}
}

// ConsumeRedis 是否从 Redis 消费。
func ConsumeRedis() bool {
	if !ProduceRedis() {
		return false
	}
	if normalizeBackend(setting.QueueSetting.Backend) == "both" {
		return normalizeConsumeFrom(setting.QueueSetting.ConsumeFrom) == "redis"
	}
	return true
}

// ConsumePulsar 是否从 Pulsar 消费。
func ConsumePulsar() bool {
	if !ProducePulsar() {
		return false
	}
	if normalizeBackend(setting.QueueSetting.Backend) == "both" {
		return normalizeConsumeFrom(setting.QueueSetting.ConsumeFrom) == "pulsar"
	}
	return true
}

func hasAsyncQueue() bool {
	if ProduceRedis() && redis.Enabled() {
		return true
	}
	if ProducePulsar() && pulsar.Enabled() {
		return true
	}
	return false
}

func combineProduceErrors(needRedis bool, redisErr error, needPulsar bool, pulsarErr error) error {
	if needRedis && redisErr != nil {
		return redisErr
	}
	if needPulsar && pulsarErr != nil {
		return pulsarErr
	}
	return nil
}
