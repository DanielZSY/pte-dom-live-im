package handler

import "testing"

func TestRandomIMSecret(t *testing.T) {
	secret, err := randomIMSecret()
	if err != nil {
		t.Fatal(err)
	}
	validateIMAlnum(t, secret, "secret")
}

func TestRandomIMSecretKeyID(t *testing.T) {
	keyID, err := randomIMSecretKeyID()
	if err != nil {
		t.Fatal(err)
	}
	validateIMAlnum(t, keyID, "key_id")
}

func validateIMAlnum(t *testing.T, text, name string) {
	t.Helper()
	if len(text) != 32 {
		t.Fatalf("%s length = %d, want 32", name, len(text))
	}
	for _, ch := range text {
		if ch >= '0' && ch <= '9' {
			continue
		}
		if ch >= 'A' && ch <= 'Z' {
			continue
		}
		if ch >= 'a' && ch <= 'z' {
			continue
		}
		t.Fatalf("%s contains non-alphanumeric char %q", name, ch)
	}
}
