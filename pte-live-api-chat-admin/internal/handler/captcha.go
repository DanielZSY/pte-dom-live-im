package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"pte_live_api_chat_admin/pkg/response"
	"pte_live_api_chat_admin/pkg/setting"
)

const (
	captchaAlphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"
	captchaLength   = 4
	captchaTTL      = 5 * time.Minute
)

var captchaColors = []string{
	"#0f172a",
	"#1d4ed8",
	"#047857",
	"#7c3aed",
	"#be123c",
	"#c2410c",
}

type captchaPayload struct {
	CodeDigest string `json:"code_digest"`
	Exp        int64  `json:"exp"`
	Nonce      string `json:"nonce"`
}

func (h *Handlers) Captcha(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}
	code := randomCaptchaCode()
	exp := time.Now().Add(captchaTTL).Unix()
	nonce := randomCaptchaCode() + randomCaptchaCode()
	payload := captchaPayload{
		CodeDigest: captchaDigest(nonce, exp, code),
		Exp:        exp,
		Nonce:      nonce,
	}
	raw, _ := json.Marshal(payload)
	body := base64.RawURLEncoding.EncodeToString(raw)
	captchaID := body + "." + tokenSignature("captcha."+body)
	response.Success(w, map[string]interface{}{
		"captcha_id":     captchaID,
		"expire_seconds": int(captchaTTL.Seconds()),
		"image":          captchaImageDataURI(code),
	})
}

func verifyCaptcha(captchaID, code string) bool {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != captchaLength {
		return false
	}
	parts := strings.Split(strings.TrimSpace(captchaID), ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	if !hmac.Equal([]byte(tokenSignature("captcha."+parts[0])), []byte(parts[1])) {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	var payload captchaPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return false
	}
	if payload.Exp <= time.Now().Unix() || payload.Nonce == "" || payload.CodeDigest == "" {
		return false
	}
	return hmac.Equal([]byte(captchaDigest(payload.Nonce, payload.Exp, code)), []byte(payload.CodeDigest))
}

func captchaDigest(nonce string, exp int64, code string) string {
	mac := hmac.New(sha256.New, []byte(setting.Auth.TokenSecret))
	_, _ = mac.Write([]byte(fmt.Sprintf("%s|%d|%s", nonce, exp, strings.ToUpper(code))))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func randomCaptchaCode() string {
	var b strings.Builder
	for i := 0; i < captchaLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(captchaAlphabet))))
		if err != nil {
			b.WriteByte(captchaAlphabet[time.Now().UnixNano()%int64(len(captchaAlphabet))])
			continue
		}
		b.WriteByte(captchaAlphabet[n.Int64()])
	}
	return b.String()
}

func randomSVGInt(max int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return time.Now().UnixNano() % max
	}
	return n.Int64()
}

func randomSVGColor() string {
	return captchaColors[randomSVGInt(int64(len(captchaColors)))]
}

func captchaImageDataURI(code string) string {
	rotations := []int{-12, 8, -7, 11}
	var noise strings.Builder
	for i := 0; i < 9; i++ {
		x1 := randomSVGInt(132)
		y1 := randomSVGInt(44)
		x2 := randomSVGInt(132)
		y2 := randomSVGInt(44)
		noise.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="%d" opacity=".34"/>`, x1, y1, x2, y2, randomSVGColor(), 1+randomSVGInt(2)))
	}
	for i := 0; i < 18; i++ {
		cx := randomSVGInt(132)
		cy := randomSVGInt(44)
		r := 1 + randomSVGInt(3)
		noise.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="%s" opacity=".25"/>`, cx, cy, r, randomSVGColor()))
	}
	for i := 0; i < 3; i++ {
		y1 := 8 + randomSVGInt(12)
		y2 := 18 + randomSVGInt(18)
		y3 := 12 + randomSVGInt(22)
		noise.WriteString(fmt.Sprintf(`<path d="M-4 %d C24 %d, 52 %d, 82 %d S118 %d, 136 %d" fill="none" stroke="%s" stroke-width="2" opacity=".42"/>`, y1, y2, y3, y1, y2, y3, randomSVGColor()))
	}
	var chars strings.Builder
	for i, r := range code {
		x := 18 + i*28
		y := 31 + int(randomSVGInt(5))
		chars.WriteString(fmt.Sprintf(`<text x="%d" y="%d" fill="%s" transform="rotate(%d %d %d)">%c</text>`, x, y, randomSVGColor(), rotations[i%len(rotations)], x, y, r))
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="132" height="44" viewBox="0 0 132 44">
<defs>
  <linearGradient id="bg" x1="0" x2="1" y1="0" y2="1">
    <stop offset="0%%" stop-color="#eff6ff"/>
    <stop offset="52%%" stop-color="#fdf2f8"/>
    <stop offset="100%%" stop-color="#ecfeff"/>
  </linearGradient>
</defs>
<rect width="132" height="44" rx="8" fill="url(#bg)"/>
%s
<g font-family="Arial, Helvetica, sans-serif" font-size="24" font-weight="800" letter-spacing="2">%s</g>
</svg>`, noise.String(), chars.String())
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}
