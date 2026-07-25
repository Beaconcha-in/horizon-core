package network

import (
	"net"
	"strings"
)

// NormalizeIPv6 نرمال‌سازی آدرس IPv6
func NormalizeIPv6(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	// اگر IPv4 باشد، به IPv6 تبدیل نمی‌کنیم
	if parsed.To4() != nil {
		return ip
	}
	return parsed.String()
}

// IsValidIPv6 بررسی اعتبار IPv6
func IsValidIPv6(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.To4() == nil
}

// ExpandIPv6 گسترش IPv6 فشرده
func ExpandIPv6(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.To4() != nil {
		return ip
	}

	parts := strings.Split(ip, "::")
	if len(parts) == 2 {
		left := strings.Split(parts[0], ":")
		right := strings.Split(parts[1], ":")
		missing := 8 - (len(left) + len(right))
		if missing < 0 {
			missing = 0
		}
		fullParts := append(left, make([]string, missing)...)
		fullParts = append(fullParts, right...)
		for i, p := range fullParts {
			if p == "" {
				fullParts[i] = "0000"
			} else {
				fullParts[i] = p
			}
		}
		return strings.Join(fullParts, ":")
	}
	return ip
}

// IPv6ToHex تبدیل IPv6 به هگز
func IPv6ToHex(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	return parsed.String()
}
