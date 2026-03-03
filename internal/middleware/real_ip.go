package middleware

import (
	"net"
	"net/http"

	"recipebox-backend-go/internal/utils"
)

func RealIP(trustedProxies []*net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if clientIP := utils.ClientIP(r, trustedProxies); clientIP != "" {
				r.RemoteAddr = clientIP
			}
			next.ServeHTTP(w, r)
		})
	}
}
