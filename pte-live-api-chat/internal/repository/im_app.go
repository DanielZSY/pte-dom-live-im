package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"pte_live_api_chat/internal/model"
)

type IMAppRepository struct {
	db *gorm.DB
}

func NewIMAppRepository(db *gorm.DB) *IMAppRepository {
	return &IMAppRepository{db: db}
}

func (r *IMAppRepository) Ready() bool {
	return r != nil && r.db != nil
}

func (r *IMAppRepository) EnsureAppForBusinessApp(ctx context.Context, appID int, merchantID uint64, name string) (*model.IMApp, *model.IMAppSecret, error) {
	if !r.Ready() {
		return nil, nil, ErrChatNotInitialized
	}
	if appID <= 0 {
		appID = 10001
	}
	imAppID, err := r.ResolveIMAppID(ctx, appID)
	if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(name) == "" {
		name = fmt.Sprintf("IM App %d", imAppID)
	}
	sdkAppID := defaultSDKAppID(imAppID)
	var app model.IMApp
	var secret model.IMAppSecret
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		app = model.IMApp{
			MerchantID:  merchantID,
			AppID:       imAppID,
			SDKAppID:    sdkAppID,
			Name:        name,
			Status:      model.IMAppStatusNormal,
			PackageCode: "free",
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "app_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"name":       gorm.Expr("IF(name = '', ?, name)", name),
				"status":     model.IMAppStatusNormal,
				"sdk_app_id": sdkAppID,
			}),
		}).Create(&app).Error; err != nil {
			return err
		}
		if err := tx.Where("app_id = ?", imAppID).First(&app).Error; err != nil {
			return err
		}
		err := tx.Where("sdk_app_id = ? AND status = ?", app.SDKAppID, model.IMSecretStatusActive).
			Order("secret_version DESC, id DESC").First(&secret).Error
		if err == nil {
			return nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		rawSecret, err := randomSecret()
		if err != nil {
			return err
		}
		keyID, err := randomIMSDKKeyID()
		if err != nil {
			return err
		}
		now := time.Now().Unix()
		secret = model.IMAppSecret{
			SDKAppID:      app.SDKAppID,
			KeyID:         keyID,
			SecretCipher:  "plain:" + rawSecret,
			SecretVersion: 1,
			Status:        model.IMSecretStatusActive,
			ActivatedAt:   now,
			CreatedBy:     "system",
		}
		return tx.Create(&secret).Error
	})
	return &app, &secret, err
}

func (r *IMAppRepository) ResolveIMAppID(ctx context.Context, businessAppID int) (int, error) {
	if !r.Ready() {
		return 0, ErrChatNotInitialized
	}
	if businessAppID <= 0 {
		businessAppID = 10001
	}
	var binding model.IMAppBinding
	err := r.db.WithContext(ctx).Where("app_id = ?", businessAppID).First(&binding).Error
	if err == nil && binding.IMAppID > 0 {
		return binding.IMAppID, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	return businessAppID, nil
}

func (r *IMAppRepository) ActiveAppAndSecret(ctx context.Context, sdkAppID string) (*model.IMApp, *model.IMAppSecret, error) {
	if !r.Ready() {
		return nil, nil, ErrChatNotInitialized
	}
	var app model.IMApp
	if err := r.db.WithContext(ctx).
		Where("sdk_app_id = ? AND status = ?", strings.TrimSpace(sdkAppID), model.IMAppStatusNormal).
		First(&app).Error; err != nil {
		return nil, nil, err
	}
	var secret model.IMAppSecret
	now := time.Now().Unix()
	if err := r.db.WithContext(ctx).
		Where("sdk_app_id = ? AND status = ? AND activated_at <= ? AND (expired_at = 0 OR expired_at > ?)", app.SDKAppID, model.IMSecretStatusActive, now, now).
		Order("secret_version DESC, id DESC").First(&secret).Error; err != nil {
		return nil, nil, err
	}
	return &app, &secret, nil
}

func (r *IMAppRepository) PackageLimitsForApp(ctx context.Context, businessAppID int) (IMQuotaLimits, error) {
	if !r.Ready() {
		return DefaultIMQuotaLimits(), ErrChatNotInitialized
	}
	return packageLimitsForApp(ctx, r.db, businessAppID)
}

func (r *IMAppRepository) LogSigIssue(ctx context.Context, row model.IMSigIssueLog) error {
	if !r.Ready() {
		return ErrChatNotInitialized
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func DecodeSecretCipher(cipher string) string {
	cipher = strings.TrimSpace(cipher)
	if strings.HasPrefix(cipher, "plain:") {
		return strings.TrimPrefix(cipher, "plain:")
	}
	return cipher
}

func defaultSDKAppID(appID int) string {
	if appID <= 0 {
		appID = 10001
	}
	return strconv.Itoa(1400000000 + appID)
}

func randomSecret() (string, error) {
	return randomIMSDKToken(32)
}

func randomIMSDKKeyID() (string, error) {
	return randomIMSDKToken(32)
}

func randomIMSDKToken(length int) (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	if length <= 0 {
		return "", nil
	}
	buf := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = alphabet[n.Int64()]
	}
	return string(buf), nil
}
