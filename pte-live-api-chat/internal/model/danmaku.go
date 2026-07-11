package model

type WxLiveDanmaku struct {
	MessageID   int64  `gorm:"column:message_id;primaryKey"`
	AppID       int    `gorm:"column:app_id"`
	LiveID      int    `gorm:"column:live_id"`
	SessionID   string `gorm:"column:session_id"`
	UserID      int    `gorm:"column:user_id"`
	NickName    string `gorm:"column:nick_name"`
	Avatar      string `gorm:"column:avatar"`
	Role        int    `gorm:"column:role"`
	Content     string `gorm:"column:content"`
	AuditStatus int    `gorm:"column:audit_status"`
	BlockType   int    `gorm:"column:block_type"`
	AuditUserID int    `gorm:"column:audit_user_id"`
	AuditTime   int64  `gorm:"column:audit_time"`
	IsBroadcast int    `gorm:"column:is_broadcast"`
	SendTime    int64  `gorm:"column:send_time"`
	Source      int    `gorm:"column:source"`
	CreateTime  int64  `gorm:"column:create_time"`
}

func (WxLiveDanmaku) TableName() string { return "pte_live_app_wx_live_danmaku" }
