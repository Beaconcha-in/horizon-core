package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// SubnetInfo اطلاعات زیرنت
type SubnetInfo struct {
	NetworkAddress string
	Broadcast      string
	FirstIP        string
	LastIP         string
	TotalIPs       int
	Netmask        string
	CIDR           int
}

// CalculateSubnet محاسبه اطلاعات زیرنت
func CalculateSubnet(cidr string) (*SubnetInfo, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	ones, bits := ipnet.Mask.Size()
	totalIPs := 1 << (bits - ones)

	ip := ipnet.IP.To4()
	if ip == nil {
		// IPv6 پشتیبانی نمی‌شود
		return nil, fmt.Errorf("IPv6 not supported yet")
	}

	// محاسبه آدرس شبکه
	network := ip.Mask(ipnet.Mask)

	// محاسبه آدرس Broadcast
	broadcast := make(net.IP, len(ipnet.IP))
	copy(broadcast, ipnet.IP)
	for i := range broadcast {
		broadcast[i] = broadcast[i] | ^ipnet.Mask[i]
	}

	// محاسبه اولین و آخرین IP
	firstIP := make(net.IP, len(ipnet.IP))
	copy(firstIP, network)
	firstIP[3]++

	lastIP := make(net.IP, len(broadcast))
	copy(lastIP, broadcast)
	lastIP[3]--

	return &SubnetInfo{
		NetworkAddress: network.String(),
		Broadcast:      broadcast.String(),
		FirstIP:        firstIP.String(),
		LastIP:         lastIP.String(),
		TotalIPs:       totalIPs,
		Netmask:        net.IP(ipnet.Mask).String(),
		CIDR:           ones,
	}, nil
}

// ExpandCIDR گسترش محدوده IPها
func ExpandCIDR(cidr string, limit int) ([]string, error) {
	info, err := CalculateSubnet(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	first := net.ParseIP(info.FirstIP).To4()
	if first == nil {
		return nil, fmt.Errorf("invalid IP")
	}

	max := info.TotalIPs - 2
	if limit > max {
		limit = max
	}
	if limit > 100 {
		limit = 100 // محدودیت برای جلوگیری از خروجی بزرگ
	}

	for i := 0; i < limit; i++ {
		ip := make(net.IP, 4)
		copy(ip, first)
		ip[3] += byte(i)
		ips = append(ips, ip.String())
	}

	if info.TotalIPs-2 > limit {
		ips = append(ips, "...")
	}

	return ips, nil
}

// IsIPInSubnet بررسی اینکه آیا IP در زیرنت است
func IsIPInSubnet(ipStr, cidr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return ipnet.Contains(ip)
}
