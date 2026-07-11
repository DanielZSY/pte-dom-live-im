package queue

import (
	"context"
	"encoding/json"
	"time"

	"pte_live_im/pkg/pulsar"
	"pte_live_im/pkg/redis"
)

const globalQueueKey = "live:queue:global"

// Message 队列消息
type Message struct {
	MessageId string `json:"messageId"`
	AppId     string `json:"appId"`
	RoomId    string `json:"roomId"`
	ClientId  string `json:"clientId"`
	UserId    string `json:"userId"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Data      string `json:"data"`
	CreatedAt int64  `json:"createdAt"`
}

// Enqueue 生产消息入队；无可用队列后端时同步 dispatch。
func Enqueue(msg Message) error {
	if msg.CreatedAt == 0 {
		msg.CreatedAt = time.Now().Unix()
	}
	if !hasAsyncQueue() {
		return dispatch(msg)
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	var redisErr, pulsarErr error
	needRedis := ProduceRedis() && redis.Enabled()
	needPulsar := ProducePulsar() && pulsar.Enabled()

	if needRedis {
		redisErr = redis.Client().LPush(context.Background(), globalQueueKey, string(raw)).Err()
	}
	if needPulsar {
		pulsarErr = pulsar.Produce(context.Background(), msg.RoomId, raw)
	}

	return combineProduceErrors(needRedis, redisErr, needPulsar, pulsarErr)
}
