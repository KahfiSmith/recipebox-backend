package utils

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP returns the caller IP, trusting forwarded headers only when the
// immediate peer is inside the configured trusted proxy ranges.
func ClientIP(r *http.Request, trustedProxies []*net.IPNet) string {
	peerIP, peerText := remoteIP(r.RemoteAddr)
	if peerIP == nil || !isTrustedProxy(peerIP, trustedProxies) {
		return peerText
	}

	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		for _, part := range strings.Split(xff, ",") {
			candidateIP, candidateText := remoteIP(strings.TrimSpace(part))
			if candidateIP != nil {
				return candidateIP.String()
			}
			if candidateText != "" {
				return candidateText
			}
		}
	}

	if xri := strings.TrimSpace(r.Header.Get("X-Real-Ip")); xri != "" {
		candidateIP, candidateText := remoteIP(xri)
		if candidateIP != nil {
			return candidateIP.String()
		}
		if candidateText != "" {
			return candidateText
		}
	}

	return peerText
}

func remoteIP(remoteAddr string) (net.IP, string) {
	remoteAddr = strings.TrimSpace(remoteAddr)
	if remoteAddr == "" {
		return nil, ""
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		remoteAddr = strings.TrimSpace(host)
	}

	ip := net.ParseIP(remoteAddr)
	if ip != nil {
		return ip, ip.String()
	}

	return nil, remoteAddr
}

func isTrustedProxy(ip net.IP, trustedProxies []*net.IPNet) bool {
	for _, network := range trustedProxies {
		if network != nil && network.Contains(ip) {
			return true
		}
	}
	return false
}
