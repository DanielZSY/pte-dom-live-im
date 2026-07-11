package handler

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestCaptchaVerify(t *testing.T) {
	code := "A7K9"
	exp := int64(4_102_444_800)
	nonce := "NONCE123"
	payload := captchaPayload{
		CodeDigest: captchaDigest(nonce, exp, code),
		Exp:        exp,
		Nonce:      nonce,
	}
	captchaID := mustCaptchaID(t, payload)
	if !verifyCaptcha(captchaID, "a7k9") {
		t.Fatalf("captcha should accept valid code case-insensitively")
	}
	if verifyCaptcha(captchaID, "B7K9") {
		t.Fatalf("captcha should reject invalid code")
	}
	if verifyCaptcha(captchaID+"x", "A7K9") {
		t.Fatalf("captcha should reject tampered id")
	}
}

func mustCaptchaID(t *testing.T, payload captchaPayload) string {
	t.Helper()
	raw, err := jsonMarshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return raw + "." + tokenSignature("captcha."+raw)
}

func jsonMarshal(payload captchaPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
