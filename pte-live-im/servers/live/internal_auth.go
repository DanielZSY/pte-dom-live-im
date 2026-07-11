package live

import (
	"net/http"
	"strings"

	"pte_live_im/pkg/setting"
)

// AuthorizeInternalAdmin 平台 admin JWT 或 service bot JWT。
func AuthorizeInternalAdmin(r *http.Request) bool {
	token := ExtractBearer(r.Header.Get("authori-zation"))
	if token == "" {
		token = ExtractBearer(r.Header.Get("authorize"))
	}
	salt := strings.TrimSpace(setting.LiveSetting.JwtSalt)
	if token == "" || salt == "" {
		return false
	}
	return ValidateAdminOrService(token, salt)
}

func ExtractBearer(headerValue string) string {
	headerValue = strings.TrimSpace(headerValue)
	if strings.HasPrefix(strings.ToLower(headerValue), "bearer ") {
		return strings.TrimSpace(headerValue[7:])
	}
	return headerValue
}

func ValidateAdminOrService(token, salt string) bool {
	for _, typ := range []string{"admin", "service"} {
		if ok, _ := validateInternalJWT(token, salt, typ); ok {
			return true
		}
	}
	return false
}
