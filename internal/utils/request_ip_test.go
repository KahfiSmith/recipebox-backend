package utils

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIPIgnoresForwardedHeadersWithoutTrustedProxy(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	req.Header.Set("X-Real-Ip", "203.0.113.11")
	req.RemoteAddr = "198.18.0.5:12345"

	if got := ClientIP(req, nil); got != "198.18.0.5" {
		t.Fatalf("expected direct peer IP, got %q", got)
	}
}

func TestClientIPUsesForwardedHeadersFromTrustedProxy(t *testing.T) {
	t.Parallel()

	_, trustedProxy, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatalf("parse trusted proxy cidr: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.10, 10.1.2.3")
	req.RemoteAddr = "10.1.2.3:12345"

	if got := ClientIP(req, []*net.IPNet{trustedProxy}); got != "203.0.113.10" {
		t.Fatalf("expected forwarded client IP, got %q", got)
	}
}
