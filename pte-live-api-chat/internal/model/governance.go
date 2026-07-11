package model

import "time"

const (
	SensitiveWordStatusDisabled = 0
	SensitiveWordStatusEnabled  = 1
)

type IMSensitiveWord struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID       int       `gorm:"column:app_id;uniqueIndex:uniq_im_sensitive_word,priority:1;index:idx_im_sensitive_word_app" json:"app_id"`
	Word        string    `gorm:"column:word;size:128;uniqueIndex:uniq_im_sensitive_word,priority:2" json:"word"`
	MatchType   string    `gorm:"column:match_type;size:16;default:contains" json:"match_type"`
	Action      string    `gorm:"column:action;size:16;default:reject" json:"action"`
	Replacement string    `gorm:"column:replacement;size:128" json:"replacement"`
	Status      int       `gorm:"column:status;default:1;index:idx_im_sensitive_word_app" json:"status"`
	HitCount    int64     `gorm:"column:hit_count" json:"hit_count"`
	CreatedBy   string    `gorm:"column:created_by;size:64" json:"created_by"`
	UpdatedBy   string    `gorm:"column:updated_by;size:64" json:"updated_by"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (IMSensitiveWord) TableName() string { return "im_sensitive_word" }

type IMSensitiveHit struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID          int       `gorm:"column:app_id;index:idx_im_sensitive_hit_app" json:"app_id"`
	WordID         uint64    `gorm:"column:word_id;index:idx_im_sensitive_hit_word" json:"word_id"`
	Word           string    `gorm:"column:word;size:128" json:"word"`
	Scene          string    `gorm:"column:scene;size:32;index:idx_im_sensitive_hit_scene" json:"scene"`
	TargetID       string    `gorm:"column:target_id;size:96;index:idx_im_sensitive_hit_scene" json:"target_id"`
	MessageID      uint64    `gorm:"column:message_id;index:idx_im_sensitive_hit_message" json:"message_id"`
	UserID         int64     `gorm:"column:user_id;index:idx_im_sensitive_hit_user" json:"user_id"`
	Action         string    `gorm:"column:action;size:16" json:"action"`
	ContentSnippet string    `gorm:"column:content_snippet;size:512" json:"content_snippet"`
	CreatedAt      time.Time `gorm:"column:created_at;index:idx_im_sensitive_hit_app" json:"created_at"`
}

func (IMSensitiveHit) TableName() string { return "im_sensitive_hit" }

type IMUserStatus struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID        int       `gorm:"column:app_id;uniqueIndex:uniq_im_user_status,priority:1;index:idx_im_user_status_status,priority:1" json:"app_id"`
	UserID       int64     `gorm:"column:user_id;uniqueIndex:uniq_im_user_status,priority:2" json:"user_id"`
	Status       int       `gorm:"column:status;default:1;index:idx_im_user_status_status,priority:2" json:"status"`
	MuteUntil    int64     `gorm:"column:mute_until" json:"mute_until"`
	DisableUntil int64     `gorm:"column:disable_until" json:"disable_until"`
	Reason       string    `gorm:"column:reason;size:255" json:"reason"`
	UpdatedBy    string    `gorm:"column:updated_by;size:64" json:"updated_by"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;index:idx_im_user_status_status,priority:3" json:"updated_at"`
}

func (IMUserStatus) TableName() string { return "im_user_status" }
