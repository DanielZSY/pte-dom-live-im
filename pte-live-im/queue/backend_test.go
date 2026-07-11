package queue

import (
	"testing"

	"pte_live_im/pkg/setting"
)

func TestBackendRouting(t *testing.T) {
	setting.QueueSetting = &setting.QueueConf{}
	setting.PulsarSetting = &setting.PulsarConf{Enabled: true}
	if ProduceRedis() || !ProducePulsar() || !ConsumePulsar() || ConsumeRedis() {
		t.Fatalf("default backend should use pulsar")
	}

	setting.QueueSetting = &setting.QueueConf{Backend: "redis", ConsumeFrom: "redis"}
	setting.PulsarSetting = &setting.PulsarConf{Enabled: true}

	if !ProduceRedis() || ProducePulsar() || ConsumePulsar() {
		t.Fatalf("redis-only backend mismatch")
	}

	setting.QueueSetting.Backend = "pulsar"
	if ProduceRedis() || !ProducePulsar() || !ConsumePulsar() || ConsumeRedis() {
		t.Fatalf("pulsar-only backend mismatch")
	}

	setting.QueueSetting.Backend = "both"
	setting.QueueSetting.ConsumeFrom = "redis"
	if !ProduceRedis() || !ProducePulsar() || !ConsumeRedis() || ConsumePulsar() {
		t.Fatalf("both+consume redis mismatch")
	}

	setting.QueueSetting.ConsumeFrom = "pulsar"
	if !ProduceRedis() || !ProducePulsar() || ConsumeRedis() || !ConsumePulsar() {
		t.Fatalf("both+consume pulsar mismatch")
	}
}
