package live

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"pte_live_im/pkg/redis"
	"pte_live_im/pkg/setting"
)

// AuthResult 鉴权结果
type AuthResult struct {
	UserID    string
	TokenType string // user | shop
	Mode      string // jwt
}

var jwtTokenTypes = []string{"user", "shop"}

// Authenticate WS token：JWT 模式，业务系统只需要按 IM 约定签发 user / shop token。
func Authenticate(token string) (AuthResult, error) {
	if setting.LiveSetting.SkipTokenValidate {
		return AuthResult{Mode: "skip"}, nil
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return AuthResult{}, errors.New("token不能为空")
	}

	for _, tokenType := range jwtTokenTypes {
		uid, err := authJWT(token, tokenType)
		if err == nil && uid != "" {
			return AuthResult{
				UserID:    uid,
				TokenType: tokenType,
				Mode:      "jwt",
			}, nil
		}
	}
	return AuthResult{}, errors.New("token无效或已过期")
}

type apiJWTClaims struct {
	Data struct {
		UID      interface{} `json:"uid"`
		Type     string      `json:"type"`
		AppID    interface{} `json:"appId"`
		AppIDAlt interface{} `json:"app_id"`
		UserRole string      `json:"userRole"`
		UserRoleAlt string   `json:"user_role"`
	} `json:"data"`
	jwt.RegisteredClaims
}

func authJWT(token, tokenType string) (string, error) {
	cfg := setting.LiveSetting
	if cfg.JwtSalt == "" {
		return "", errors.New("jwt not configured")
	}
	if redis.Enabled() && isJWTBlacklisted(token, tokenType) {
		return "", errors.New("jwt blacklisted")
	}

	key := cfg.JwtSalt + tokenType
	if key == tokenType {
		return "", errors.New("jwt salt empty")
	}

	claims := &apiJWTClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(key), nil
	}, jwt.WithLeeway(time.Duration(cfg.JwtLeeway)*time.Second))
	if err != nil || !parsed.Valid {
		return "", err
	}
	if claims.Data.Type != "" && claims.Data.Type != tokenType {
		return "", errors.New("jwt type mismatch")
	}
	uid := formatUID(claims.Data.UID)
	if uid == "" {
		return "", errors.New("jwt uid empty")
	}
	return uid, nil
}

func isJWTBlacklisted(token, tokenType string) bool {
	sum := sha1.Sum([]byte(token))
	key := fmt.Sprintf("auth:%s:blacklist:%s", tokenType, hex.EncodeToString(sum[:]))
	n, err := redis.Client().Exists(ctx(), key).Result()
	return err == nil && n > 0
}

func formatUID(v interface{}) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ValidateToken 兼容旧调用
func ValidateToken(token, userId string) error {
	res, err := Authenticate(token)
	if err != nil {
		return err
	}
	if userId != "" && res.UserID != "" && res.UserID != userId {
		return errors.New("token与userId不匹配")
	}
	return nil
}
