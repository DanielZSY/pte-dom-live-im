package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"pte_live_api_chat/internal/model"
)

const (
	DefaultMaxUserGroups            = 10000
	DefaultMaxGroupMembers          = 100000
	DefaultMaxLiveRoomOnline        = 1000000
	DefaultMaxVoiceRoomOnline       = 1000000
	DefaultMaxConnections           = 1000000
	DefaultMaxConcurrentConnections = 100000
)

type IMQuotaLimits struct {
	MaxUserGroups            int `json:"max_user_groups"`
	MaxGroupMembers          int `json:"max_group_members"`
	MaxLiveRoomOnline        int `json:"max_live_room_online"`
	MaxVoiceRoomOnline       int `json:"max_voice_room_online"`
	MaxConnections           int `json:"max_connections"`
	MaxConcurrentConnections int `json:"max_concurrent_connections"`
}

func DefaultIMQuotaLimits() IMQuotaLimits {
	return IMQuotaLimits{
		MaxUserGroups:            DefaultMaxUserGroups,
		MaxGroupMembers:          DefaultMaxGroupMembers,
		MaxLiveRoomOnline:        DefaultMaxLiveRoomOnline,
		MaxVoiceRoomOnline:       DefaultMaxVoiceRoomOnline,
		MaxConnections:           DefaultMaxConnections,
		MaxConcurrentConnections: DefaultMaxConcurrentConnections,
	}
}

func packageLimitsForApp(ctx context.Context, db *gorm.DB, businessAppID int) (IMQuotaLimits, error) {
	limits := DefaultIMQuotaLimits()
	if db == nil {
		return limits, nil
	}
	if businessAppID <= 0 {
		businessAppID = 10001
	}
	imAppID := businessAppID
	var binding model.IMAppBinding
	err := db.WithContext(ctx).Where("app_id = ?", businessAppID).First(&binding).Error
	if err == nil && binding.IMAppID > 0 {
		imAppID = binding.IMAppID
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return limits, err
	}

	var app model.IMApp
	err = db.WithContext(ctx).Where("app_id = ? AND status = ?", imAppID, model.IMAppStatusNormal).First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return limits, nil
	}
	if err != nil {
		return limits, err
	}
	packageCode := strings.TrimSpace(app.PackageCode)
	if packageCode == "" {
		packageCode = "free"
	}

	var pkg model.IMPackage
	err = db.WithContext(ctx).Where("code = ? AND status = ?", packageCode, 1).First(&pkg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return limits, nil
	}
	if err != nil {
		return limits, err
	}
	return normalizeLimits(IMQuotaLimits{
		MaxUserGroups:            pkg.MaxUserGroups,
		MaxGroupMembers:          pkg.MaxGroupMembers,
		MaxLiveRoomOnline:        pkg.MaxLiveRoomOnline,
		MaxVoiceRoomOnline:       pkg.MaxVoiceRoomOnline,
		MaxConnections:           pkg.MaxConnections,
		MaxConcurrentConnections: pkg.MaxConcurrentConnections,
	}), nil
}

func normalizeLimits(limits IMQuotaLimits) IMQuotaLimits {
	defaults := DefaultIMQuotaLimits()
	if limits.MaxUserGroups <= 0 {
		limits.MaxUserGroups = defaults.MaxUserGroups
	}
	if limits.MaxGroupMembers <= 0 {
		limits.MaxGroupMembers = defaults.MaxGroupMembers
	}
	if limits.MaxLiveRoomOnline <= 0 {
		limits.MaxLiveRoomOnline = defaults.MaxLiveRoomOnline
	}
	if limits.MaxVoiceRoomOnline <= 0 {
		limits.MaxVoiceRoomOnline = defaults.MaxVoiceRoomOnline
	}
	if limits.MaxConnections <= 0 {
		limits.MaxConnections = defaults.MaxConnections
	}
	if limits.MaxConcurrentConnections <= 0 {
		limits.MaxConcurrentConnections = defaults.MaxConcurrentConnections
	}
	return limits
}

func quotaExceededError(name string, current, limit int64) error {
	return fmt.Errorf("%s已达到上限：当前 %d，上限 %d", name, current, limit)
}
