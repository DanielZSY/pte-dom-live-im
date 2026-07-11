package appid

import (
	"net/http"
	"testing"
)

func TestTokenFromHTTP(t *testing.T) {
	t.Run("authori-zation Bearer", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "http://localhost/ws", nil)
		r.Header.Set("authori-zation", "Bearer jwt-from-auth")
		if got := TokenFromHTTP(r); got != "jwt-from-auth" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("Token header", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "http://localhost/ws", nil)
		r.Header.Set(HeaderToken, "jwt-from-token-header")
		if got := TokenFromHTTP(r); got != "jwt-from-token-header" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("query token", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "http://localhost/ws?token=jwt-from-query", nil)
		if got := TokenFromHTTP(r); got != "jwt-from-query" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("authori-zation wins over Token header", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "http://localhost/ws", nil)
		r.Header.Set("authori-zation", "Bearer jwt-auth")
		r.Header.Set(HeaderToken, "jwt-header")
		if got := TokenFromHTTP(r); got != "jwt-auth" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("Token header wins over query", func(t *testing.T) {
		r, _ := http.NewRequest(http.MethodGet, "http://localhost/ws?token=jwt-query", nil)
		r.Header.Set(HeaderToken, "jwt-header")
		if got := TokenFromHTTP(r); got != "jwt-header" {
			t.Fatalf("got %q", got)
		}
	})
}
