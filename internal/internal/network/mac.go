package network

import (
	"strings"
)

// OUIMap دیتابیس آفلاین OUI (نمونه)
var ouidb = map[string]string{
	"00:1A:2B": "Cisco Systems",
	"00:14:22": "Dell Inc.",
	"00:50:56": "VMware",
	"00:0C:29": "VMware",
	"00:1C:42": "Cisco Systems",
	"00:1E:8C": "Apple Inc.",
	"00:25:00": "Apple Inc.",
	"00:23:DF": "Intel Corporation",
	"00:15:5D": "Microsoft Corporation",
	"00:0F:20": "Nokia",
	"00:0D:93": "Samsung Electronics",
	"00:1F:3C": "Hewlett-Packard",
	"00:1F:29": "Dell Inc.",
	"00:1B:21": "Hewlett-Packard",
	"00:1E:4F": "Dell Inc.",
	"00:1D:09": "Apple Inc.",
	"00:1D:4F": "Apple Inc.",
	"00:0E:C6": "Samsung Electronics",
	"00:0E:07": "Samsung Electronics",
	"00:12:3F": "Samsung Electronics",
}

// LookupOUI جستجوی OUI در دیتابیس آفلاین
func LookupOUI(mac string) string {
	mac = strings.ToUpper(mac)
	mac = strings.ReplaceAll(mac, "-", ":")
	mac = strings.ReplaceAll(mac, ".", ":")

	// استخراج OUI (3 بایت اول)
	parts := strings.Split(mac, ":")
	if len(parts) < 3 {
		return "MAC address format invalid"
	}

	oui := strings.Join(parts[:3], ":")

	if vendor, ok := ouidb[oui]; ok {
		return vendor
	}
	return "Vendor not found in offline database"
}

// IsValidMAC بررسی اعتبار MAC
func IsValidMAC(mac string) bool {
	mac = strings.ReplaceAll(mac, "-", ":")
	mac = strings.ReplaceAll(mac, ".", ":")
	parts := strings.Split(mac, ":")
	if len(parts) != 6 {
		return false
	}
	for _, p := range parts {
		if len(p) != 2 {
			return false
		}
		for _, c := range p {
			if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
				return false
			}
		}
	}
	return true
}
