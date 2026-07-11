package service

import (
	"testing"
	"time"
)

func TestUserSigSignAndVerify(t *testing.T) {
	secret := "test-secret"
	payload := userSigPayload{
		SDKAppID:   "1400010001",
		AppID:      "10001",
		Identifier: "user_10086",
		UserID:     "10086",
		UserType:   "user",
		Expire:     60,
		Time:       time.Now().Unix(),
		Nonce:      "nonce",
		KeyID:      "kid_1",
	}

	token, err := signUserSig(payload, secret)
	if err != nil {
		t.Fatalf("signUserSig() error = %v", err)
	}
	got, err := verifyUserSig(token, secret)
	if err != nil {
		t.Fatalf("verifyUserSig() error = %v", err)
	}
	if got.SDKAppID != payload.SDKAppID || got.Identifier != payload.Identifier || got.UserID != payload.UserID {
		t.Fatalf("verify payload mismatch: got %+v want %+v", got, payload)
	}

	if _, err := verifyUserSig(token+"x", secret); err == nil {
		t.Fatalf("verifyUserSig() should reject tampered token")
	}
}

func TestUserSigExpiredPayload(t *testing.T) {
	payload := userSigPayload{Time: time.Now().Add(-2 * time.Hour).Unix(), Expire: 60}
	if payload.Time <= 0 || payload.Expire <= 0 || payload.Time+payload.Expire >= time.Now().Unix() {
		t.Fatalf("test payload should be expired")
	}
}
