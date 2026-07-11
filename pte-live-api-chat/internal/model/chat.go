package model

import "time"

const (
	ConversationTypeSingle = "single"
	ConversationTypeGroup  = "group"

	ConversationStatusNormal   = 1
	ConversationStatusDisabled = 2

	MemberRoleOwner  = 1
	MemberRoleAdmin  = 2
	MemberRoleMember = 3

	MessageStatusNormal   = 1
	MessageStatusRecalled = 2
	MessageStatusDeleted  = 3
)

type ChatConversation struct {
	ID                  uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID               int       `gorm:"column:app_id;index:idx_chat_conv_app_updated;uniqueIndex:uniq_chat_conv_single,priority:1" json:"app_id"`
	Type                string    `gorm:"column:type;size:16;index:idx_chat_conv_type" json:"type"`
	SingleKey           string    `gorm:"column:single_key;size:96;uniqueIndex:uniq_chat_conv_single,priority:2" json:"single_key"`
	GroupID             string    `gorm:"column:group_id;size:64;index:idx_chat_conv_group" json:"group_id"`
	Title               string    `gorm:"column:title;size:128" json:"title"`
	Avatar              string    `gorm:"column:avatar;size:512" json:"avatar"`
	Status              int       `gorm:"column:status;default:1" json:"status"`
	LastMessageID       uint64    `gorm:"column:last_message_id" json:"last_message_id"`
	LastMessageSeq      int64     `gorm:"column:last_message_seq" json:"last_message_seq"`
	LastMessageSnapshot string    `gorm:"column:last_message_snapshot;size:1024" json:"last_message_snapshot"`
	LastMessageAt       int64     `gorm:"column:last_message_at" json:"last_message_at"`
	CreatedAt           time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;index:idx_chat_conv_app_updated,priority:2" json:"updated_at"`
}

func (ChatConversation) TableName() string { return "chat_conversation" }

type ChatMember struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID          int       `gorm:"column:app_id;uniqueIndex:uniq_chat_member,priority:1" json:"app_id"`
	ConversationID uint64    `gorm:"column:conversation_id;uniqueIndex:uniq_chat_member,priority:2;index:idx_chat_member_conv" json:"conversation_id"`
	UserID         int64     `gorm:"column:user_id;uniqueIndex:uniq_chat_member,priority:3;index:idx_chat_member_user" json:"user_id"`
	Role           int       `gorm:"column:role;default:3" json:"role"`
	Alias          string    `gorm:"column:alias;size:128" json:"alias"`
	MuteUntil      int64     `gorm:"column:mute_until" json:"mute_until"`
	LastReadSeq    int64     `gorm:"column:last_read_seq" json:"last_read_seq"`
	UnreadCount    int64     `gorm:"column:unread_count" json:"unread_count"`
	JoinedAt       int64     `gorm:"column:joined_at" json:"joined_at"`
	DeletedAtUnix  int64     `gorm:"column:deleted_at" json:"deleted_at"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ChatMember) TableName() string { return "chat_member" }

type ChatMessage struct {
	ID               uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID            int       `gorm:"column:app_id;uniqueIndex:uniq_chat_msg_client,priority:1;index:idx_chat_msg_conv_seq,priority:1" json:"app_id"`
	ConversationID   uint64    `gorm:"column:conversation_id;index:idx_chat_msg_conv_seq,priority:2" json:"conversation_id"`
	ConversationType string    `gorm:"column:conversation_type;size:16" json:"conversation_type"`
	SenderID         int64     `gorm:"column:sender_id;uniqueIndex:uniq_chat_msg_client,priority:2;index:idx_chat_msg_sender" json:"sender_id"`
	ClientMsgID      string    `gorm:"column:client_msg_id;size:96;uniqueIndex:uniq_chat_msg_client,priority:3" json:"client_msg_id"`
	MsgType          string    `gorm:"column:msg_type;size:32" json:"msg_type"`
	Content          string    `gorm:"column:content;size:4096" json:"content"`
	Payload          string    `gorm:"column:payload;type:json" json:"payload"`
	QuoteMessageID   uint64    `gorm:"column:quote_message_id" json:"quote_message_id"`
	QuoteSnapshot    string    `gorm:"column:quote_snapshot;size:2048" json:"quote_snapshot"`
	Status           int       `gorm:"column:status;default:1" json:"status"`
	Seq              int64     `gorm:"column:seq;index:idx_chat_msg_conv_seq,priority:3" json:"seq"`
	SentAt           int64     `gorm:"column:sent_at" json:"sent_at"`
	RecalledAt       int64     `gorm:"column:recalled_at" json:"recalled_at"`
	DeletedAtUnix    int64     `gorm:"column:deleted_at" json:"deleted_at"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ChatMessage) TableName() string { return "chat_message" }

type ChatMessageUserState struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID          int       `gorm:"column:app_id;uniqueIndex:uniq_chat_msg_user_state,priority:1" json:"app_id"`
	MessageID      uint64    `gorm:"column:message_id;uniqueIndex:uniq_chat_msg_user_state,priority:2" json:"message_id"`
	ConversationID uint64    `gorm:"column:conversation_id;index:idx_chat_msg_state_user" json:"conversation_id"`
	UserID         int64     `gorm:"column:user_id;uniqueIndex:uniq_chat_msg_user_state,priority:3;index:idx_chat_msg_state_user" json:"user_id"`
	IsDeleted      int       `gorm:"column:is_deleted" json:"is_deleted"`
	DeletedAtUnix  int64     `gorm:"column:deleted_at" json:"deleted_at"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ChatMessageUserState) TableName() string { return "chat_message_user_state" }

type ChatMessageReceipt struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID          int       `gorm:"column:app_id;uniqueIndex:uniq_chat_msg_receipt,priority:1;index:idx_chat_msg_receipt_user" json:"app_id"`
	MessageID      uint64    `gorm:"column:message_id;uniqueIndex:uniq_chat_msg_receipt,priority:2;index:idx_chat_msg_receipt_msg" json:"message_id"`
	ConversationID uint64    `gorm:"column:conversation_id;index:idx_chat_msg_receipt_user" json:"conversation_id"`
	UserID         int64     `gorm:"column:user_id;uniqueIndex:uniq_chat_msg_receipt,priority:3;index:idx_chat_msg_receipt_user" json:"user_id"`
	DeviceID       string    `gorm:"column:device_id;size:96;uniqueIndex:uniq_chat_msg_receipt,priority:4" json:"device_id"`
	DeliveredAt    int64     `gorm:"column:delivered_at" json:"delivered_at"`
	ReadAt         int64     `gorm:"column:read_at" json:"read_at"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ChatMessageReceipt) TableName() string { return "chat_message_receipt" }

type ChatOutbox struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID       int       `gorm:"column:app_id;index:idx_chat_outbox_status" json:"app_id"`
	EventID     string    `gorm:"column:event_id;size:96;uniqueIndex:uniq_chat_outbox_event" json:"event_id"`
	EventType   string    `gorm:"column:event_type;size:64" json:"event_type"`
	Payload     string    `gorm:"column:payload;type:json" json:"payload"`
	Status      int       `gorm:"column:status;index:idx_chat_outbox_status" json:"status"`
	Retry       int       `gorm:"column:retry" json:"retry"`
	NextAt      int64     `gorm:"column:next_at" json:"next_at"`
	LockedUntil int64     `gorm:"column:locked_until" json:"locked_until"`
	LastError   string    `gorm:"column:last_error;size:512" json:"last_error"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ChatOutbox) TableName() string { return "chat_outbox" }
