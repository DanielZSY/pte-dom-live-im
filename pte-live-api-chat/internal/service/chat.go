package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"pte_live_api_chat/internal/model"
	"pte_live_api_chat/internal/repository"
)

type ChatService struct {
	repo *repository.ChatRepository
}

func NewChatService(repo *repository.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) Ready() bool {
	return s != nil && s.repo != nil && s.repo.Ready()
}

func (s *ChatService) EnsureSchema(ctx context.Context) error {
	if !s.Ready() {
		return repository.ErrChatNotInitialized
	}
	return s.repo.EnsureSchema(ctx)
}

type OpenSingleRequest struct {
	AppID        int   `json:"app_id"`
	UserID       int64 `json:"user_id"`
	PeerUserID   int64 `json:"peer_user_id"`
	TargetUserID int64 `json:"target_user_id"`
}

type CreateGroupRequest struct {
	AppID     int     `json:"app_id"`
	OwnerID   int64   `json:"owner_id"`
	UserID    int64   `json:"user_id"`
	Title     string  `json:"title"`
	Avatar    string  `json:"avatar"`
	MemberIDs []int64 `json:"member_ids"`
	MemberID  []int64 `json:"member_id"`
}

type ConversationListRequest struct {
	AppID    int   `json:"app_id"`
	UserID   int64 `json:"user_id"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type ConversationDetailRequest struct {
	AppID          int    `json:"app_id"`
	UserID         int64  `json:"user_id"`
	ConversationID uint64 `json:"conversation_id"`
}

type ConversationReadRequest struct {
	AppID          int    `json:"app_id"`
	UserID         int64  `json:"user_id"`
	ConversationID uint64 `json:"conversation_id"`
	Seq            int64  `json:"seq"`
}

type MessageSendRequest struct {
	AppID          int    `json:"app_id"`
	ConversationID uint64 `json:"conversation_id"`
	SenderID       int64  `json:"sender_id"`
	UserID         int64  `json:"user_id"`
	ClientMsgID    string `json:"client_msg_id"`
	MsgType        string `json:"msg_type"`
	Type           string `json:"type"`
	Content        string `json:"content"`
	Payload        string `json:"payload"`
	QuoteMessageID uint64 `json:"quote_message_id"`
}

type MessageHistoryRequest struct {
	AppID          int    `json:"app_id"`
	UserID         int64  `json:"user_id"`
	ConversationID uint64 `json:"conversation_id"`
	BeforeSeq      int64  `json:"before_seq"`
	Limit          int    `json:"limit"`
}

type MessageSyncRequest struct {
	AppID          int    `json:"app_id"`
	UserID         int64  `json:"user_id"`
	ConversationID uint64 `json:"conversation_id"`
	AfterSeq       int64  `json:"after_seq"`
	Limit          int    `json:"limit"`
}

type MessageAckRequest struct {
	AppID          int      `json:"app_id"`
	UserID         int64    `json:"user_id"`
	ConversationID uint64   `json:"conversation_id"`
	MessageIDs     []uint64 `json:"message_ids"`
	Seq            int64    `json:"seq"`
	AckType        string   `json:"ack_type"`
	DeviceID       string   `json:"device_id"`
}

type MessageActionRequest struct {
	AppID      int    `json:"app_id"`
	UserID     int64  `json:"user_id"`
	OperatorID int64  `json:"operator_id"`
	MessageID  uint64 `json:"message_id"`
}

type MemberListRequest struct {
	AppID          int    `json:"app_id"`
	UserID         int64  `json:"user_id"`
	ConversationID uint64 `json:"conversation_id"`
}

type MemberAddRequest struct {
	AppID          int     `json:"app_id"`
	UserID         int64   `json:"user_id"`
	OperatorID     int64   `json:"operator_id"`
	ConversationID uint64  `json:"conversation_id"`
	MemberIDs      []int64 `json:"member_ids"`
}

type MemberRemoveRequest struct {
	AppID          int    `json:"app_id"`
	UserID         int64  `json:"user_id"`
	OperatorID     int64  `json:"operator_id"`
	ConversationID uint64 `json:"conversation_id"`
	MemberID       int64  `json:"member_id"`
}

type ConversationView struct {
	ID                  uint64       `json:"id"`
	AppID               int          `json:"app_id"`
	Type                string       `json:"type"`
	SingleKey           string       `json:"single_key,omitempty"`
	GroupID             string       `json:"group_id,omitempty"`
	Title               string       `json:"title"`
	Avatar              string       `json:"avatar"`
	Status              int          `json:"status"`
	LastMessageID       uint64       `json:"last_message_id"`
	LastMessageSeq      int64        `json:"last_message_seq"`
	LastMessageSnapshot string       `json:"last_message_snapshot"`
	LastMessageAt       int64        `json:"last_message_at"`
	Role                int          `json:"role,omitempty"`
	LastReadSeq         int64        `json:"last_read_seq,omitempty"`
	UnreadCount         int64        `json:"unread_count,omitempty"`
	Members             []MemberView `json:"members,omitempty"`
}

type MemberView struct {
	UserID      int64  `json:"user_id"`
	Role        int    `json:"role"`
	Alias       string `json:"alias"`
	MuteUntil   int64  `json:"mute_until"`
	LastReadSeq int64  `json:"last_read_seq"`
	UnreadCount int64  `json:"unread_count"`
	JoinedAt    int64  `json:"joined_at"`
}

type MessageView struct {
	MessageID        uint64 `json:"message_id"`
	AppID            int    `json:"app_id"`
	ConversationID   uint64 `json:"conversation_id"`
	ConversationType string `json:"conversation_type"`
	SenderID         int64  `json:"sender_id"`
	ClientMsgID      string `json:"client_msg_id"`
	MsgType          string `json:"msg_type"`
	Content          string `json:"content"`
	Payload          string `json:"payload"`
	QuoteMessageID   uint64 `json:"quote_message_id"`
	QuoteSnapshot    string `json:"quote_snapshot"`
	Status           int    `json:"status"`
	Seq              int64  `json:"seq"`
	SentAt           int64  `json:"sent_at"`
	RecalledAt       int64  `json:"recalled_at,omitempty"`
}

type ConversationListResult struct {
	Total int64              `json:"total"`
	List  []ConversationView `json:"list"`
}

type MessageListResult struct {
	List []MessageView `json:"list"`
}

type MessageSyncResult struct {
	List    []MessageView `json:"list"`
	NextSeq int64         `json:"next_seq"`
	HasMore bool          `json:"has_more"`
}

type OKResult struct {
	OK bool `json:"ok"`
}

func (s *ChatService) OpenSingle(ctx context.Context, appID int, req OpenSingleRequest) (*ConversationView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	userID := req.UserID
	peerID := firstPositive64(req.PeerUserID, req.TargetUserID)
	if userID <= 0 || peerID <= 0 || userID == peerID {
		return nil, errors.New("单聊用户参数错误")
	}
	conv, err := s.repo.OpenSingle(ctx, appID, userID, peerID)
	if err != nil {
		return nil, err
	}
	row, members, err := s.repo.ConversationDetail(ctx, appID, conv.ID, userID)
	if err != nil {
		return nil, err
	}
	view := conversationView(*row)
	view.Members = memberViews(members)
	return &view, nil
}

func (s *ChatService) CreateGroup(ctx context.Context, appID int, req CreateGroupRequest) (*ConversationView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	ownerID := firstPositive64(req.OwnerID, req.UserID)
	if ownerID <= 0 {
		return nil, errors.New("缺少群主")
	}
	memberIDs := req.MemberIDs
	if len(memberIDs) == 0 {
		memberIDs = req.MemberID
	}
	conv, err := s.repo.CreateGroup(ctx, appID, ownerID, req.Title, req.Avatar, memberIDs)
	if err != nil {
		return nil, err
	}
	row, members, err := s.repo.ConversationDetail(ctx, appID, conv.ID, ownerID)
	if err != nil {
		return nil, err
	}
	view := conversationView(*row)
	view.Members = memberViews(members)
	return &view, nil
}

func (s *ChatService) ListConversations(ctx context.Context, appID int, req ConversationListRequest) (*ConversationListResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	if req.UserID <= 0 {
		return nil, errors.New("缺少 user_id")
	}
	rows, total, err := s.repo.ListConversations(ctx, appID, req.UserID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	out := make([]ConversationView, 0, len(rows))
	for _, row := range rows {
		out = append(out, conversationView(row))
	}
	return &ConversationListResult{Total: total, List: out}, nil
}

func (s *ChatService) ConversationDetail(ctx context.Context, appID int, req ConversationDetailRequest) (*ConversationView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	if req.UserID <= 0 || req.ConversationID == 0 {
		return nil, errors.New("参数错误")
	}
	row, members, err := s.repo.ConversationDetail(ctx, appID, req.ConversationID, req.UserID)
	if err != nil {
		return nil, err
	}
	view := conversationView(*row)
	view.Members = memberViews(members)
	return &view, nil
}

func (s *ChatService) MarkRead(ctx context.Context, appID int, req ConversationReadRequest) (*OKResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	if req.UserID <= 0 || req.ConversationID == 0 {
		return nil, errors.New("参数错误")
	}
	if err := s.repo.MarkRead(ctx, appID, req.ConversationID, req.UserID, req.Seq); err != nil {
		return nil, err
	}
	return &OKResult{OK: true}, nil
}

func (s *ChatService) SendMessage(ctx context.Context, appID int, req MessageSendRequest) (*MessageView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	senderID := firstPositive64(req.SenderID, req.UserID)
	msgType := strings.TrimSpace(req.MsgType)
	if msgType == "" {
		msgType = strings.TrimSpace(req.Type)
	}
	if msgType == "" {
		msgType = "text"
	}
	content := strings.TrimSpace(req.Content)
	if req.ConversationID == 0 || senderID <= 0 {
		return nil, errors.New("消息参数错误")
	}
	if msgType == "text" && content == "" {
		return nil, errors.New("消息内容不能为空")
	}
	if utf8.RuneCountInString(content) > 2000 {
		return nil, errors.New("消息内容最多 2000 字")
	}
	var moderationReq *repository.ContentModerationRequest
	var moderationResult *repository.ContentModerationResult
	if msgType == "text" {
		modReq := repository.ContentModerationRequest{
			AppID:    appID,
			Scene:    "chat",
			TargetID: fmt.Sprintf("%d", req.ConversationID),
			UserID:   senderID,
			Content:  content,
		}
		moderation, err := s.repo.ModerateContent(ctx, modReq)
		if err != nil {
			return nil, err
		}
		if moderation.Blocked {
			if err := s.repo.RecordContentModerationHits(ctx, modReq, moderation, 0); err != nil {
				return nil, err
			}
			return nil, errors.New("消息包含敏感词")
		}
		moderationReq = &modReq
		moderationResult = moderation
		content = moderation.Content
	}
	clientMsgID := strings.TrimSpace(req.ClientMsgID)
	if clientMsgID == "" {
		clientMsgID = fmt.Sprintf("srv-%d-%d", senderID, time.Now().UnixNano())
	}
	payload := strings.TrimSpace(req.Payload)
	if payload == "" {
		payload = "{}"
	}
	msg, err := s.repo.SendMessage(ctx, repository.SendMessageParams{
		AppID:             appID,
		ConversationID:    req.ConversationID,
		SenderID:          senderID,
		ClientMsgID:       clientMsgID,
		MsgType:           msgType,
		Content:           content,
		Payload:           payload,
		QuoteMessageID:    req.QuoteMessageID,
		ModerationRequest: moderationReq,
		ModerationResult:  moderationResult,
	})
	if err != nil {
		return nil, err
	}
	view := messageView(*msg)
	return &view, nil
}

func (s *ChatService) History(ctx context.Context, appID int, req MessageHistoryRequest) (*MessageListResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	if req.UserID <= 0 || req.ConversationID == 0 {
		return nil, errors.New("参数错误")
	}
	rows, err := s.repo.History(ctx, appID, req.ConversationID, req.UserID, req.BeforeSeq, req.Limit)
	if err != nil {
		return nil, err
	}
	out := make([]MessageView, 0, len(rows))
	for _, row := range rows {
		out = append(out, messageView(row))
	}
	return &MessageListResult{List: out}, nil
}

func (s *ChatService) SyncMessages(ctx context.Context, appID int, req MessageSyncRequest) (*MessageSyncResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	if req.UserID <= 0 || req.ConversationID == 0 {
		return nil, errors.New("参数错误")
	}
	rows, hasMore, err := s.repo.SyncMessages(ctx, appID, req.ConversationID, req.UserID, req.AfterSeq, req.Limit)
	if err != nil {
		return nil, err
	}
	out := make([]MessageView, 0, len(rows))
	nextSeq := req.AfterSeq
	for _, row := range rows {
		out = append(out, messageView(row))
		if row.Seq > nextSeq {
			nextSeq = row.Seq
		}
	}
	return &MessageSyncResult{List: out, NextSeq: nextSeq, HasMore: hasMore}, nil
}

func (s *ChatService) AckMessages(ctx context.Context, appID int, req MessageAckRequest) (*OKResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	ackType := strings.TrimSpace(req.AckType)
	if ackType == "" {
		ackType = "delivered"
	}
	if ackType != "delivered" && ackType != "read" {
		return nil, errors.New("ack_type 只支持 delivered/read")
	}
	if req.UserID <= 0 || req.ConversationID == 0 {
		return nil, errors.New("参数错误")
	}
	if len(req.MessageIDs) == 0 && req.Seq <= 0 {
		return nil, errors.New("缺少 message_ids 或 seq")
	}
	if err := s.repo.AckMessages(ctx, repository.AckMessageParams{
		AppID:          appID,
		ConversationID: req.ConversationID,
		UserID:         req.UserID,
		MessageIDs:     req.MessageIDs,
		Seq:            req.Seq,
		AckType:        ackType,
		DeviceID:       req.DeviceID,
	}); err != nil {
		return nil, err
	}
	return &OKResult{OK: true}, nil
}

func (s *ChatService) RecallMessage(ctx context.Context, appID int, req MessageActionRequest) (*MessageView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	operatorID := firstPositive64(req.OperatorID, req.UserID)
	if operatorID <= 0 || req.MessageID == 0 {
		return nil, errors.New("参数错误")
	}
	msg, err := s.repo.RecallMessage(ctx, appID, req.MessageID, operatorID)
	if err != nil {
		return nil, err
	}
	view := messageView(*msg)
	return &view, nil
}

func (s *ChatService) DeleteMessage(ctx context.Context, appID int, req MessageActionRequest) (*OKResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	userID := firstPositive64(req.UserID, req.OperatorID)
	if userID <= 0 || req.MessageID == 0 {
		return nil, errors.New("参数错误")
	}
	if err := s.repo.DeleteMessageForUser(ctx, appID, req.MessageID, userID); err != nil {
		return nil, err
	}
	return &OKResult{OK: true}, nil
}

func (s *ChatService) ListMembers(ctx context.Context, appID int, req MemberListRequest) ([]MemberView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	if req.UserID <= 0 || req.ConversationID == 0 {
		return nil, errors.New("参数错误")
	}
	if _, _, err := s.repo.ConversationDetail(ctx, appID, req.ConversationID, req.UserID); err != nil {
		return nil, err
	}
	members, err := s.repo.ListMembers(ctx, appID, req.ConversationID)
	return memberViews(members), err
}

func (s *ChatService) AddMembers(ctx context.Context, appID int, req MemberAddRequest) (*OKResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	operatorID := firstPositive64(req.OperatorID, req.UserID)
	if operatorID <= 0 || req.ConversationID == 0 {
		return nil, errors.New("参数错误")
	}
	if err := s.repo.AddGroupMembers(ctx, appID, req.ConversationID, operatorID, req.MemberIDs); err != nil {
		return nil, err
	}
	return &OKResult{OK: true}, nil
}

func (s *ChatService) RemoveMember(ctx context.Context, appID int, req MemberRemoveRequest) (*OKResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	operatorID := firstPositive64(req.OperatorID, req.UserID)
	if operatorID <= 0 || req.ConversationID == 0 || req.MemberID <= 0 {
		return nil, errors.New("参数错误")
	}
	if err := s.repo.RemoveGroupMember(ctx, appID, req.ConversationID, operatorID, req.MemberID); err != nil {
		return nil, err
	}
	return &OKResult{OK: true}, nil
}

func conversationView(row repository.ConversationRow) ConversationView {
	return ConversationView{
		ID:                  row.ID,
		AppID:               row.AppID,
		Type:                row.Type,
		SingleKey:           row.SingleKey,
		GroupID:             row.GroupID,
		Title:               row.Title,
		Avatar:              row.Avatar,
		Status:              row.Status,
		LastMessageID:       row.LastMessageID,
		LastMessageSeq:      row.LastMessageSeq,
		LastMessageSnapshot: row.LastMessageSnapshot,
		LastMessageAt:       row.LastMessageAt,
		Role:                row.Role,
		LastReadSeq:         row.LastReadSeq,
		UnreadCount:         row.UnreadCount,
	}
}

func memberViews(rows []model.ChatMember) []MemberView {
	out := make([]MemberView, 0, len(rows))
	for _, row := range rows {
		out = append(out, MemberView{
			UserID:      row.UserID,
			Role:        row.Role,
			Alias:       row.Alias,
			MuteUntil:   row.MuteUntil,
			LastReadSeq: row.LastReadSeq,
			UnreadCount: row.UnreadCount,
			JoinedAt:    row.JoinedAt,
		})
	}
	return out
}

func messageView(row model.ChatMessage) MessageView {
	return MessageView{
		MessageID:        row.ID,
		AppID:            row.AppID,
		ConversationID:   row.ConversationID,
		ConversationType: row.ConversationType,
		SenderID:         row.SenderID,
		ClientMsgID:      row.ClientMsgID,
		MsgType:          row.MsgType,
		Content:          row.Content,
		Payload:          row.Payload,
		QuoteMessageID:   row.QuoteMessageID,
		QuoteSnapshot:    row.QuoteSnapshot,
		Status:           row.Status,
		Seq:              row.Seq,
		SentAt:           row.SentAt,
		RecalledAt:       row.RecalledAt,
	}
}

func firstPositive(a, b int) int {
	if a > 0 {
		return a
	}
	return b
}

func firstPositive64(values ...int64) int64 {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 0
}
