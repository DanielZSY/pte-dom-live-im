package routers

import (
	"net/http"

	"pte_live_im/pkg/cors"
)

// CORSMiddleware handles browser cross-origin preflight and response headers.
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return cors.Middleware(next)
}
