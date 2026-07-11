package queue

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"pte_live_im/define/livecode"
	"pte_live_im/pkg/pulsar"
	"pte_live_im/pkg/redis"
	"pte_live_im/pkg/setting"
	"pte_live_im/servers"
	"pte_live_im/servers/live"
	"pte_live_im/tools/util"
)

var processedIds = make(map[string]struct{})
var processedIdsMu sync.Mutex
var processedIdOrder []string

const maxProcessedIds = 10000

// StartWorkers 按 queue.backend / queue.consumeFrom 启动 Redis 与/或 Pulsar 消费者。
func StartWorkers() {
	n := setting.LiveSetting.QueueWorkers
	if n <= 0 {
		n = 2
	}

	started := 0
	if ConsumeRedis() && redis.Enabled() {
		for i := 0; i < n; i++ {
			go redisWorker(i)
		}
		log.Infof("live queue redis workers started: %d (backend=%s)", n, setting.QueueSetting.Backend)
		started += n
	} else if ConsumeRedis() {
		log.Warn("queue.consumeFrom=redis 但 Redis 未连接，跳过 Redis 消费者")
	}

	if ConsumePulsar() && pulsar.Enabled() {
		for i := 0; i < n; i++ {
			go pulsarWorker(i)
		}
		log.Infof("live queue pulsar workers started: %d (backend=%s)", n, setting.QueueSetting.Backend)
		started += n
	} else if ConsumePulsar() {
		log.Warn("queue.consumeFrom=pulsar 但 Pulsar 未连接，跳过 Pulsar 消费者")
	}

	if started == 0 {
		log.Warn("未启动任何队列消费者（检查 queue.backend / redis / pulsar 配置）")
	}
}

func redisWorker(id int) {
	for {
		result, err := redis.Client().BRPop(context.Background(), 3*time.Second, globalQueueKey).Result()
		if err != nil {
			continue
		}
		if len(result) < 2 {
			continue
		}
		var msg Message
		if json.Unmarshal([]byte(result[1]), &msg) != nil {
			continue
		}
		if err := handleMessage(msg); err != nil {
			log.WithFields(log.Fields{"worker": id, "backend": "redis", "messageId": msg.MessageId, "err": err}).Error("queue dispatch failed")
		}
	}
}

func pulsarWorker(id int) {
	consumer, err := pulsar.NewSharedConsumer()
	if err != nil {
		log.WithFields(log.Fields{"worker": id, "err": err}).Error("pulsar consumer 创建失败")
		return
	}
	defer consumer.Close()

	ctx := context.Background()
	for {
		pMsg, err := consumer.Receive(ctx)
		if err != nil {
			log.WithFields(log.Fields{"worker": id, "err": err}).Error("pulsar receive failed")
			time.Sleep(time.Second)
			continue
		}

		var msg Message
		if json.Unmarshal(pMsg.Payload(), &msg) != nil {
			consumer.Nack(pMsg)
			continue
		}
		if err := handleMessage(msg); err != nil {
			log.WithFields(log.Fields{"worker": id, "backend": "pulsar", "messageId": msg.MessageId, "err": err}).Error("queue dispatch failed")
			consumer.Nack(pMsg)
			continue
		}
		consumer.Ack(pMsg)
	}
}

func handleMessage(msg Message) error {
	return dispatch(msg)
}

func dispatch(msg Message) error {
	if markProcessed(msg.MessageId) {
		return nil
	}

	groupName := livecode.GroupName(msg.RoomId)
	data := msg.Data

	switch msg.Code {
	case livecode.Danmaku:
		role := live.ParseDanmakuRole(json.RawMessage(msg.Data))
		if live.IsMuted(msg.AppId, msg.RoomId, msg.UserId, role) {
			log.WithFields(log.Fields{
				"appId":     msg.AppId,
				"roomId":    msg.RoomId,
				"userId":    msg.UserId,
				"role":      role,
				"messageId": msg.MessageId,
				"code":      msg.Code,
			}).Warn("live queue danmaku dropped by mute state")
			return nil
		}
		// api-live 已审核弹幕（ClientId 为空）直接广播，跳过 IM 侧二次审核
		if msg.ClientId != "" {
			cfg := live.GetConfig(msg.AppId, msg.RoomId)
			if cfg.DanmakuAudit && live.IsDanmakuSubjectToAudit(role) {
				live.AddPendingDanmaku(msg.AppId, msg.RoomId, msg.MessageId, msg.Data)
				log.WithFields(log.Fields{
					"appId":     msg.AppId,
					"roomId":    msg.RoomId,
					"userId":    msg.UserId,
					"role":      role,
					"messageId": msg.MessageId,
					"code":      msg.Code,
				}).Info("live queue danmaku moved to pending audit")
				return nil
			}
		}
		log.WithFields(log.Fields{
			"appId":     msg.AppId,
			"roomId":    msg.RoomId,
			"userId":    msg.UserId,
			"role":      role,
			"messageId": msg.MessageId,
			"code":      msg.Code,
			"groupName": groupName,
		}).Info("live queue danmaku dispatch to group")
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.GiftSend:
		var gift live.GiftRecord
		if json.Unmarshal([]byte(msg.Data), &gift) == nil {
			gift.UserId = msg.UserId
			live.AddGift(msg.AppId, msg.RoomId, gift)
		}
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.LinkMicApply:
		var item live.LinkMicItem
		if json.Unmarshal([]byte(msg.Data), &item) == nil {
			item.UserId = msg.UserId
			live.AddLinkMicApply(msg.AppId, msg.RoomId, item)
		}
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.MuteAll:
		live.SetMuteAll(msg.AppId, msg.RoomId, true)
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.UnmuteAll:
		live.SetMuteAll(msg.AppId, msg.RoomId, false)
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.MuteUser:
		var payload struct {
			UserId string `json:"userId"`
		}
		if json.Unmarshal([]byte(msg.Data), &payload) == nil && payload.UserId != "" {
			live.MuteUser(msg.AppId, msg.RoomId, payload.UserId)
		}
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.UnmuteUser:
		var payload struct {
			UserId string `json:"userId"`
		}
		if json.Unmarshal([]byte(msg.Data), &payload) == nil && payload.UserId != "" {
			live.UnmuteUser(msg.AppId, msg.RoomId, payload.UserId)
		}
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.DanmakuAuditOn:
		live.SetConfigField(msg.AppId, msg.RoomId, "danmakuAudit", "true")
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.DanmakuAuditOff:
		live.SetConfigField(msg.AppId, msg.RoomId, "danmakuAudit", "false")
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.LinkMicAgree, livecode.LinkMicReject:
		var payload struct {
			UserId string `json:"userId"`
		}
		if json.Unmarshal([]byte(msg.Data), &payload) == nil && payload.UserId != "" {
			live.RemoveLinkMicApply(msg.AppId, msg.RoomId, payload.UserId)
		}
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.KickUser:
		var payload struct {
			UserId string `json:"userId"`
		}
		if json.Unmarshal([]byte(msg.Data), &payload) == nil && payload.UserId != "" {
			live.AppKickUser(msg.AppId, payload.UserId)
			live.KickUser(msg.AppId, msg.RoomId, payload.UserId)
			kickClientsInApp(msg.AppId, payload.UserId)
			kickClientsInRoom(msg.AppId, msg.RoomId, payload.UserId)
		}
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	case livecode.ConfigChange:
		applyConfigChange(msg.AppId, msg.RoomId, msg.Data)
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)

	default:
		if msg.MessageId == "" {
			msg.MessageId = util.GenUUID()
		}
		servers.SendMessage2Group(msg.AppId, msg.UserId, groupName, msg.Code, msg.Msg, &data)
	}
	return nil
}

func markProcessed(messageId string) bool {
	if messageId == "" {
		return false
	}
	processedIdsMu.Lock()
	defer processedIdsMu.Unlock()
	if _, ok := processedIds[messageId]; ok {
		return true
	}
	processedIds[messageId] = struct{}{}
	processedIdOrder = append(processedIdOrder, messageId)
	if len(processedIdOrder) > maxProcessedIds {
		evict := processedIdOrder[0]
		processedIdOrder = processedIdOrder[1:]
		delete(processedIds, evict)
	}
	return false
}

func applyConfigChange(appId, roomId, data string) {
	var cfg map[string]interface{}
	if json.Unmarshal([]byte(data), &cfg) != nil {
		return
	}
	for k, v := range cfg {
		switch val := v.(type) {
		case bool:
			live.SetConfigField(appId, roomId, k, boolToStr(val))
		case string:
			live.SetConfigField(appId, roomId, k, val)
		}
	}
}

func boolToStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func kickClientsInRoom(appId, roomId, userId string) {
	groupKey := util.GenGroupKey(appId, livecode.GroupName(roomId))
	clientIds := servers.Manager.GetGroupClientList(groupKey)
	for _, cid := range clientIds {
		if client, err := servers.Manager.GetByClientId(cid); err == nil && client.UserId == userId {
			servers.CloseClient(cid, appId)
		}
	}
}

func kickClientsInApp(appId, userId string) {
	clientIds := servers.Manager.GetSystemClientList(appId)
	for _, cid := range clientIds {
		if client, err := servers.Manager.GetByClientId(cid); err == nil && client.UserId == userId {
			servers.CloseClient(cid, appId)
		}
	}
}
