package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pte_live_api_chat/internal/model"
	"pte_live_api_chat/internal/mq"
	"pte_live_api_chat/internal/repository"
	"pte_live_api_chat/pkg/setting"
)

const (
	chatDeliverCodeMessage       = 20001
	chatDeliverCodeRecall        = 20002
	chatDeliverCodeDelete        = 20003
	chatDeliverCodeRead          = 20004
	chatDeliverCodeMemberAdded   = 20005
	chatDeliverCodeMemberRemoved = 20006
	chatDeliverCodeDelivered     = 20007
	chatDeliverCodeReadReceipt   = 20008
	chatDeliverCodeGovernance    = 20009
)

type OutboxWorker struct {
	repo      *repository.ChatRepository
	client    *http.Client
	publisher *mq.PulsarPublisher
}

func NewOutboxWorker(repo *repository.ChatRepository) *OutboxWorker {
	var publisher *mq.PulsarPublisher
	if deliverViaPulsar() {
		p, err := mq.NewPulsarPublisher(setting.IM.PulsarServiceURL, setting.IM.PulsarTopic)
		if err != nil {
			log.Printf("api-chat pulsar publisher disabled: %v", err)
		} else {
			publisher = p
			log.Printf("api-chat pulsar publisher ready topic=%s", setting.IM.PulsarTopic)
		}
	}
	return &OutboxWorker{
		repo:      repo,
		publisher: publisher,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	if w == nil || w.repo == nil || !setting.IM.OutboxEnabled {
		return
	}
	workers := setting.IM.OutboxWorkers
	if workers <= 0 {
		workers = 2
	}
	for i := 0; i < workers; i++ {
		go w.loop(ctx, i)
	}
	log.Printf("api-chat outbox workers started: %d", workers)
}

func (w *OutboxWorker) loop(ctx context.Context, id int) {
	interval := time.Duration(setting.IM.OutboxInterval) * time.Second
	if interval <= 0 {
		interval = 2 * time.Second
	}
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if err := w.consumeOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("api-chat outbox worker %d: %v", id, err)
			}
			timer.Reset(interval)
		}
	}
}

func (w *OutboxWorker) consumeOnce(ctx context.Context) error {
	rows, err := w.repo.ClaimOutbox(ctx, setting.IM.OutboxBatchSize, int64(setting.IM.OutboxLockTTL))
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := w.deliver(ctx, row); err != nil {
			retry := row.Retry + 1
			nextAt := time.Now().Unix() + retryDelay(retry)
			dead := setting.IM.OutboxMaxRetries > 0 && retry >= setting.IM.OutboxMaxRetries
			if dead {
				nextAt = 0
			}
			if markErr := w.repo.MarkOutboxFailed(ctx, row.ID, retry, nextAt, err.Error(), dead); markErr != nil {
				return markErr
			}
			continue
		}
		if err := w.repo.MarkOutboxSent(ctx, row.ID); err != nil {
			return err
		}
	}
	return nil
}

func (w *OutboxWorker) deliver(ctx context.Context, row model.ChatOutbox) error {
	sceneEventID, err := outboxSceneEventID(row)
	if err != nil {
		return err
	}
	if sceneEventID > 0 {
		return w.deliverSceneEvent(ctx, row, sceneEventID)
	}
	if isGenericChatEvent(row.EventType) {
		return w.deliverGenericChatEvent(ctx, row)
	}
	messageID, err := outboxMessageID(row)
	if err != nil {
		return err
	}
	msg, recipients, err := w.repo.MessageWithRecipients(ctx, row.AppID, messageID)
	if err != nil {
		return err
	}
	if len(recipients) == 0 {
		return nil
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"event_id":   row.EventID,
		"event_type": row.EventType,
		"message":    messageView(*msg),
	})
	reqBody := chatDeliverBody(row.AppID, recipients, msg.SenderID, deliverCode(row.EventType), row.EventType, string(payload))
	return w.deliverToBackend(ctx, row, reqBody, "im deliver")
}

func (w *OutboxWorker) deliverGenericChatEvent(ctx context.Context, row model.ChatOutbox) error {
	var payload struct {
		MessageID      uint64   `json:"message_id"`
		ConversationID uint64   `json:"conversation_id"`
		UserID         int64    `json:"user_id"`
		OperatorID     int64    `json:"operator_id"`
		Operator       string   `json:"operator"`
		MemberIDs      []int64  `json:"member_ids"`
		MessageIDs     []uint64 `json:"message_ids"`
		Seq            int64    `json:"seq"`
		DeviceID       string   `json:"device_id"`
		AckType        string   `json:"ack_type"`
		Status         int      `json:"status"`
		MuteUntil      int64    `json:"mute_until"`
		Role           int      `json:"role"`
		Reason         string   `json:"reason"`
	}
	if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
		return err
	}
	if payload.ConversationID == 0 {
		return errors.New("outbox payload missing conversation_id")
	}
	recipients, err := w.genericRecipients(ctx, row.AppID, row.EventType, payload.ConversationID, payload.UserID)
	if err != nil {
		return err
	}
	if len(recipients) == 0 {
		return nil
	}
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"event_id":        row.EventID,
		"event_type":      row.EventType,
		"message_id":      payload.MessageID,
		"conversation_id": payload.ConversationID,
		"user_id":         payload.UserID,
		"operator_id":     payload.OperatorID,
		"operator":        payload.Operator,
		"member_ids":      payload.MemberIDs,
		"message_ids":     payload.MessageIDs,
		"seq":             payload.Seq,
		"device_id":       payload.DeviceID,
		"ack_type":        payload.AckType,
		"status":          payload.Status,
		"mute_until":      payload.MuteUntil,
		"role":            payload.Role,
		"reason":          payload.Reason,
	})
	senderID := payload.OperatorID
	if senderID <= 0 {
		senderID = payload.UserID
	}
	reqBody := chatDeliverBody(row.AppID, recipients, senderID, deliverCode(row.EventType), row.EventType, string(eventPayload))
	return w.deliverToBackend(ctx, row, reqBody, "im deliver")
}

func (w *OutboxWorker) genericRecipients(ctx context.Context, appID int, eventType string, conversationID uint64, userID int64) ([]int64, error) {
	if eventType == "chat.message.deleted" {
		if userID <= 0 {
			return nil, errors.New("outbox payload missing user_id")
		}
		return []int64{userID}, nil
	}
	return w.repo.ConversationRecipients(ctx, appID, conversationID)
}

func (w *OutboxWorker) postIMRaw(ctx context.Context, url string, appID int, raw []byte, label string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AppId", strconv.Itoa(appID))
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s http %d: %s", label, resp.StatusCode, string(body))
	}
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &ret); err != nil {
		return err
	}
	if ret.Code != 0 {
		return fmt.Errorf("%s failed: %s", label, ret.Msg)
	}
	return nil
}

func (w *OutboxWorker) deliverSceneEvent(ctx context.Context, row model.ChatOutbox, eventID uint64) error {
	event, err := w.repo.SceneEvent(ctx, row.AppID, eventID)
	if err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"event_id":   row.EventID,
		"event_type": event.EventType,
		"scene":      event,
	})
	reqBody := map[string]interface{}{
		"app_id":     strconv.Itoa(row.AppID),
		"groupName":  event.GroupName,
		"sendUserId": strconv.FormatInt(event.ActorID, 10),
		"code":       event.Code,
		"msg":        event.EventType,
		"data":       string(payload),
	}
	return w.deliverToBackend(ctx, row, reqBody, "im scene deliver")
}

func (w *OutboxWorker) deliverToBackend(ctx context.Context, row model.ChatOutbox, reqBody map[string]interface{}, label string) error {
	raw, _ := json.Marshal(reqBody)
	var errs []string
	if deliverViaPulsar() {
		if w.publisher == nil {
			errs = append(errs, "pulsar publisher unavailable")
		} else if err := w.publisher.Publish(ctx, outboxPartitionKey(row, reqBody), raw); err != nil {
			errs = append(errs, "pulsar: "+err.Error())
		}
	}
	if deliverViaHTTP() {
		path := setting.IM.DeliverPath
		if _, ok := reqBody["groupName"]; ok {
			path = "/api/send_to_group"
		}
		if err := w.postIMRaw(ctx, setting.IM.HTTPURL+path, row.AppID, raw, label); err != nil {
			errs = append(errs, "http: "+err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func outboxMessageID(row model.ChatOutbox) (uint64, error) {
	var payload struct {
		MessageID uint64 `json:"message_id"`
	}
	if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
		return 0, err
	}
	if payload.MessageID == 0 {
		return 0, errors.New("outbox payload missing message_id")
	}
	return payload.MessageID, nil
}

func outboxSceneEventID(row model.ChatOutbox) (uint64, error) {
	var payload struct {
		SceneEventID uint64 `json:"scene_event_id"`
	}
	if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
		return 0, err
	}
	return payload.SceneEventID, nil
}

func deliverCode(eventType string) int {
	switch eventType {
	case "chat.message.recalled":
		return chatDeliverCodeRecall
	case "chat.message.deleted", "chat.message.deleted_all":
		return chatDeliverCodeDelete
	case "chat.conversation.read":
		return chatDeliverCodeRead
	case "chat.member.added":
		return chatDeliverCodeMemberAdded
	case "chat.member.removed":
		return chatDeliverCodeMemberRemoved
	case "chat.message.delivered":
		return chatDeliverCodeDelivered
	case "chat.message.read":
		return chatDeliverCodeReadReceipt
	case "chat.conversation.disabled", "chat.conversation.enabled", "chat.member.muted", "chat.member.unmuted", "chat.member.role_updated":
		return chatDeliverCodeGovernance
	default:
		return chatDeliverCodeMessage
	}
}

func isGenericChatEvent(eventType string) bool {
	switch eventType {
	case "chat.message.deleted", "chat.message.deleted_all", "chat.conversation.read", "chat.member.added", "chat.member.removed", "chat.message.delivered", "chat.message.read", "chat.conversation.disabled", "chat.conversation.enabled", "chat.member.muted", "chat.member.unmuted", "chat.member.role_updated":
		return true
	default:
		return false
	}
}

func chatDeliverBody(appID int, recipients []int64, senderID int64, code int, eventType string, payload string) map[string]interface{} {
	userIDs := make([]string, 0, len(recipients))
	for _, uid := range recipients {
		if uid > 0 {
			userIDs = append(userIDs, strconv.FormatInt(uid, 10))
		}
	}
	return map[string]interface{}{
		"app_id":     strconv.Itoa(appID),
		"user_ids":   userIDs,
		"sendUserId": strconv.FormatInt(senderID, 10),
		"code":       code,
		"msg":        eventType,
		"data":       payload,
	}
}

func deliverViaHTTP() bool {
	return setting.IM.DeliverBackend == "http" || setting.IM.DeliverBackend == "both"
}

func deliverViaPulsar() bool {
	return setting.IM.DeliverBackend == "pulsar" || setting.IM.DeliverBackend == "both"
}

func outboxPartitionKey(row model.ChatOutbox, reqBody map[string]interface{}) string {
	if v, ok := reqBody["groupName"]; ok && strings.TrimSpace(fmt.Sprint(v)) != "" {
		return strings.TrimSpace(fmt.Sprint(v))
	}
	var payload struct {
		ConversationID uint64 `json:"conversation_id"`
	}
	if json.Unmarshal([]byte(row.Payload), &payload) == nil && payload.ConversationID > 0 {
		return strconv.FormatUint(payload.ConversationID, 10)
	}
	return strconv.Itoa(row.AppID)
}

func retryDelay(retry int) int64 {
	if retry <= 0 {
		return 1
	}
	delay := int64(math.Pow(2, float64(retry)))
	if delay > 300 {
		delay = 300
	}
	return delay
}
