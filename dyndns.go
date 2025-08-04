package dyndns

import (
	"fmt"
	"net"
	"strings"
)

type Domain interface {
	Add(h string, ip string) error
	Delete(h string, ip string) error
	Update(h string, ip string) error
}

func NewAWS(name, zoneID, accessKey, accessSecret string, ttl int64) Domain {
	return DomainAWS{
		Name:         name,
		ZoneID:       zoneID,
		AccessKey:    accessKey,
		AccessSecret: accessSecret,
		TTL:          ttl,
	}
}

// Do not use -> Not Implemeted yet
func NewGAE(name, zoneID, accessKey, accessSecret string, ttl int64) Domain {
	return DomainGAE{
		Name:         name,
		ZoneID:       zoneID,
		AccessKey:    accessKey,
		AccessSecret: accessSecret,
		TTL:          ttl,
	}
}

// rsType determines the DNS record type ("A" or "AAAA") based on the IP address.
func rsType(ipStr string) (string, error) {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipStr)
	}
	if ip.To4() != nil {
		return "A", nil
	}
	return "AAAA", nil
}

// validateHostname checks if the hostname is valid.
func validateHostname(h string) error {
	if h == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	// Simple validation: Only letters, numbers, dots, and hyphens allowed
	for _, c := range h {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '.' || c == '-') {
			return fmt.Errorf("invalid hostname: %s contains invalid characters", h)
		}
	}
	return nil
}
