package mq

import (
	"context"
	"errors"
	"strings"
	"sync"

	pulsar "github.com/apache/pulsar-client-go/pulsar"
)

var ErrDisabled = errors.New("pulsar publisher disabled")

type PulsarPublisher struct {
	mu       sync.RWMutex
	client   pulsar.Client
	producer pulsar.Producer
}

func NewPulsarPublisher(serviceURL string, topic string) (*PulsarPublisher, error) {
	serviceURL = strings.TrimSpace(serviceURL)
	topic = strings.TrimSpace(topic)
	if serviceURL == "" || topic == "" {
		return nil, ErrDisabled
	}
	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: serviceURL})
	if err != nil {
		return nil, err
	}
	producer, err := client.CreateProducer(pulsar.ProducerOptions{Topic: topic})
	if err != nil {
		client.Close()
		return nil, err
	}
	return &PulsarPublisher{client: client, producer: producer}, nil
}

func (p *PulsarPublisher) Publish(ctx context.Context, key string, payload []byte) error {
	if p == nil {
		return ErrDisabled
	}
	p.mu.RLock()
	producer := p.producer
	p.mu.RUnlock()
	if producer == nil {
		return ErrDisabled
	}
	_, err := producer.Send(ctx, &pulsar.ProducerMessage{Key: key, Payload: payload})
	return err
}

func (p *PulsarPublisher) Close() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.producer != nil {
		p.producer.Close()
		p.producer = nil
	}
	if p.client != nil {
		p.client.Close()
		p.client = nil
	}
}
