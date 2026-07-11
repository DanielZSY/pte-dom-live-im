package live

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	platformAppID = 10000
	serviceBotUID = 1
)

func validateInternalJWT(token, salt, tokenType string) (bool, string) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(salt + tokenType), nil
	})
	if err != nil || !parsed.Valid {
		return false, ""
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return false, ""
	}
	data, _ := claims["data"].(map[string]interface{})
	claimType := stringField(data, "type")
	if claimType == "" || claimType != tokenType {
		return false, ""
	}
	userRole := stringField(data, "userRole")
	appID := intField(data, "appId")
	uid := intField(data, "uid")
	switch tokenType {
	case "admin":
		return userRole == "admin" && appID == platformAppID && uid > 0, userRole
	case "service":
		return userRole == "bot" && appID == platformAppID && uid == serviceBotUID, userRole
	default:
		return false, ""
	}
}

func stringField(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	v, ok := data[key]
	if !ok {
		return ""
	}
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func intField(data map[string]interface{}, key string) int {
	if data == nil {
		return 0
	}
	v, ok := data[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}
