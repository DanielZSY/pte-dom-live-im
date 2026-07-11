package pulsar

import (
	"context"
	"strings"
	"sync"

	pulsar "github.com/apache/pulsar-client-go/pulsar"
	log "github.com/sirupsen/logrus"
	"pte_live_im/pkg/setting"
)

var (
	mu       sync.RWMutex
	client   pulsar.Client
	producer pulsar.Producer
	ready    bool
)

// Setup 连接 Pulsar 并创建 IM 事件 Producer（queue.backend 为 pulsar/both 时调用）。
func Setup() error {
	mu.Lock()
	defer mu.Unlock()

	if ready {
		return nil
	}
	if !setting.PulsarSetting.Enabled {
		return nil
	}
	url := setting.PulsarSetting.ServiceURL
	if url == "" {
		log.Warn("pulsar.enabled 但未配置 serviceURL，跳过连接")
		return nil
	}

	c, err := pulsar.NewClient(pulsar.ClientOptions{URL: url})
	if err != nil {
		return err
	}

	topic := setting.PulsarSetting.Topic
	p, err := c.CreateProducer(pulsar.ProducerOptions{Topic: topic})
	if err != nil {
		c.Close()
		return err
	}

	client = c
	producer = p
	ready = true
	log.Infof("Pulsar 已连接 topic=%s", topic)
	return nil
}

// Enabled 配置开启且 Producer 已就绪。
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return ready
}

// Produce 发送队列消息；partitionKey 通常为 roomId。
func Produce(ctx context.Context, partitionKey string, payload []byte) error {
	mu.RLock()
	p := producer
	mu.RUnlock()
	if p == nil {
		return errNotReady
	}
	_, err := p.Send(ctx, &pulsar.ProducerMessage{
		Key:     partitionKey,
		Payload: payload,
	})
	return err
}

// NewSharedConsumer 创建 Shared 订阅消费者（多 IM 节点水平扩展）。
func NewSharedConsumer() (pulsar.Consumer, error) {
	mu.RLock()
	c := client
	mu.RUnlock()
	if c == nil {
		return nil, errNotReady
	}
	return c.Subscribe(pulsar.ConsumerOptions{
		Topic:            setting.PulsarSetting.Topic,
		SubscriptionName: setting.PulsarSetting.Subscription,
		Type:             pulsar.Shared,
	})
}

// NewChatConsumer 创建 chat-events 订阅消费者。集群模式下每个节点使用独立订阅，
// 保证所有节点都收到事件，再只投递本机连接。
func NewChatConsumer() (pulsar.Consumer, error) {
	mu.RLock()
	c := client
	mu.RUnlock()
	if c == nil {
		return nil, errNotReady
	}
	subscription := setting.PulsarSetting.ChatSubscription
	if setting.CommonSetting.Cluster && setting.GlobalSetting.LocalHost != "" {
		subscription = subscription + "-" + sanitizeSubscription(setting.GlobalSetting.LocalHost)
	}
	return c.Subscribe(pulsar.ConsumerOptions{
		Topic:            setting.PulsarSetting.ChatTopic,
		SubscriptionName: subscription,
		Type:             pulsar.Shared,
	})
}

// Close 关闭连接。
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if producer != nil {
		producer.Close()
		producer = nil
	}
	if client != nil {
		client.Close()
		client = nil
	}
	ready = false
}

func sanitizeSubscription(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "node"
	}
	replacer := strings.NewReplacer(".", "-", ":", "-", "/", "-", "_", "-")
	return replacer.Replace(raw)
}
