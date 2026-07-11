package queue

import (
	"context"
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
	"pte_live_im/pkg/pulsar"
	"pte_live_im/pkg/setting"
	"pte_live_im/servers"
)

type chatEvent struct {
	AppID      string   `json:"app_id"`
	AppID2     string   `json:"appId"`
	UserIDs    []string `json:"user_ids"`
	UserIDs2   []string `json:"userIds"`
	GroupName  string   `json:"groupName"`
	SendUserID string   `json:"sendUserId"`
	Code       int      `json:"code"`
	Msg        string   `json:"msg"`
	Data       string   `json:"data"`
}

func StartChatWorkers() {
	if !pulsar.Enabled() || setting.PulsarSetting.ChatTopic == "" {
		return
	}
	workers := setting.PulsarSetting.ChatWorkers
	if workers <= 0 {
		workers = 2
	}
	for i := 0; i < workers; i++ {
		go chatPulsarWorker(i)
	}
	log.Infof("chat-events pulsar workers started: %d topic=%s", workers, setting.PulsarSetting.ChatTopic)
}

func chatPulsarWorker(id int) {
	consumer, err := pulsar.NewChatConsumer()
	if err != nil {
		log.WithFields(log.Fields{"worker": id, "err": err}).Error("chat pulsar consumer 创建失败")
		return
	}
	defer consumer.Close()
	ctx := context.Background()
	for {
		msg, err := consumer.Receive(ctx)
		if err != nil {
			log.WithFields(log.Fields{"worker": id, "err": err}).Error("chat pulsar receive failed")
			time.Sleep(time.Second)
			continue
		}
		var event chatEvent
		if err := json.Unmarshal(msg.Payload(), &event); err != nil {
			consumer.Nack(msg)
			continue
		}
		if err := dispatchChatEvent(event); err != nil {
			log.WithFields(log.Fields{"worker": id, "err": err}).Error("chat event dispatch failed")
			consumer.Nack(msg)
			continue
		}
		consumer.Ack(msg)
	}
}

func dispatchChatEvent(event chatEvent) error {
	appID := event.AppID
	if appID == "" {
		appID = event.AppID2
	}
	if appID == "" {
		return nil
	}
	if event.Code == 0 {
		event.Code = 20001
	}
	data := event.Data
	if event.GroupName != "" {
		servers.Manager.SendMessage2LocalGroup(appID, "", event.SendUserID, event.GroupName, event.Code, event.Msg, &data)
		return nil
	}
	userIDs := event.UserIDs
	if len(userIDs) == 0 {
		userIDs = event.UserIDs2
	}
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		servers.SendMessage2User(appID, userID, event.SendUserID, event.Code, event.Msg, &data)
	}
	return nil
}
