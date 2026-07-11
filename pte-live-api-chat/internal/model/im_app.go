package model

import "time"

const (
	IMAppStatusNormal   = 1
	IMAppStatusDisabled = 2

	IMSecretStatusActive   = 1
	IMSecretStatusDisabled = 2
	IMSecretStatusRotated  = 3
)

type IMApp struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MerchantID  uint64    `gorm:"column:merchant_id;index:idx_im_app_merchant" json:"merchant_id"`
	AppID       int       `gorm:"column:app_id;uniqueIndex:uniq_im_app_app;index:idx_im_app_app" json:"app_id"`
	SDKAppID    string    `gorm:"column:sdk_app_id;size:32;uniqueIndex:uniq_im_app_sdk" json:"sdk_app_id"`
	Name        string    `gorm:"column:name;size:128" json:"name"`
	Status      int       `gorm:"column:status;default:1;index:idx_im_app_status" json:"status"`
	PackageCode string    `gorm:"column:package_code;size:64" json:"package_code"`
	Remark      string    `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (IMApp) TableName() string { return "im_app" }

type IMPackage struct {
	ID                       uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code                     string    `gorm:"column:code;size:64;uniqueIndex:uniq_im_package_code" json:"code"`
	Name                     string    `gorm:"column:name;size:128" json:"name"`
	MonthlyPrice             float64   `gorm:"column:monthly_price" json:"monthly_price"`
	YearlyPrice              float64   `gorm:"column:yearly_price" json:"yearly_price"`
	MaxUserGroups            int       `gorm:"column:max_user_groups;default:10000" json:"max_user_groups"`
	MaxGroupMembers          int       `gorm:"column:max_group_members;default:100000" json:"max_group_members"`
	MaxLiveRoomOnline        int       `gorm:"column:max_live_room_online;default:1000000" json:"max_live_room_online"`
	MaxVoiceRoomOnline       int       `gorm:"column:max_voice_room_online;default:1000000" json:"max_voice_room_online"`
	MaxConnections           int       `gorm:"column:max_connections;default:1000000" json:"max_connections"`
	MaxConcurrentConnections int       `gorm:"column:max_concurrent_connections;default:100000" json:"max_concurrent_connections"`
	Status                   int       `gorm:"column:status;default:1;index:idx_im_package_status" json:"status"`
	Sort                     int       `gorm:"column:sort;default:100;index:idx_im_package_sort" json:"sort"`
	Remark                   string    `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt                time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (IMPackage) TableName() string { return "im_package" }

type IMAppBinding struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID     int       `gorm:"column:app_id;uniqueIndex:uniq_im_app_binding_app;index:idx_im_app_binding_app" json:"app_id"`
	IMAppID   int       `gorm:"column:im_app_id;index:idx_im_app_binding_im_app" json:"im_app_id"`
	CreatedBy string    `gorm:"column:created_by;size:64" json:"created_by"`
	UpdatedBy string    `gorm:"column:updated_by;size:64" json:"updated_by"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (IMAppBinding) TableName() string { return "im_app_binding" }

type IMAppSecret struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SDKAppID      string    `gorm:"column:sdk_app_id;size:32;index:idx_im_secret_sdk" json:"sdk_app_id"`
	KeyID         string    `gorm:"column:key_id;size:32;uniqueIndex:uniq_im_secret_key" json:"key_id"`
	SecretCipher  string    `gorm:"column:secret_cipher;size:1024" json:"secret_cipher"`
	SecretVersion int       `gorm:"column:secret_version;default:1" json:"secret_version"`
	Status        int       `gorm:"column:status;default:1;index:idx_im_secret_status" json:"status"`
	ActivatedAt   int64     `gorm:"column:activated_at" json:"activated_at"`
	ExpiredAt     int64     `gorm:"column:expired_at" json:"expired_at"`
	CreatedBy     string    `gorm:"column:created_by;size:64" json:"created_by"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (IMAppSecret) TableName() string { return "im_app_secret" }

type IMSigIssueLog struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID      int       `gorm:"column:app_id;index:idx_im_sig_app" json:"app_id"`
	SDKAppID   string    `gorm:"column:sdk_app_id;size:32;index:idx_im_sig_sdk" json:"sdk_app_id"`
	Identifier string    `gorm:"column:identifier;size:96;index:idx_im_sig_identifier" json:"identifier"`
	KeyID      string    `gorm:"column:key_id;size:32" json:"key_id"`
	UserType   string    `gorm:"column:user_type;size:32" json:"user_type"`
	DeviceID   string    `gorm:"column:device_id;size:96" json:"device_id"`
	Platform   string    `gorm:"column:platform;size:32" json:"platform"`
	Scene      string    `gorm:"column:scene;size:32" json:"scene"`
	ExpireAt   int64     `gorm:"column:expire_at" json:"expire_at"`
	IP         string    `gorm:"column:ip;size:64" json:"ip"`
	CreatedAt  time.Time `gorm:"column:created_at;index:idx_im_sig_created" json:"created_at"`
}

func (IMSigIssueLog) TableName() string { return "im_sig_issue_log" }
