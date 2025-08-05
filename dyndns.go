package dyndns

import (
	"fmt"
	"net"
	"strings"
)

// Domain defines the interface for DNS operations.
type Domain interface {
	Add(h string, ip string) error
	Delete(h string, ip string, types []string) error
	Update(h string, ip string) error
	List(hostname string) ([]Record, error)
	DeleteRecords(h string, types []string, force bool) error
}

// Record represents a DNS record.
type Record struct {
	Name   string
	Type   string
	TTL    int64
	Values []string
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

// contains checks if a slice contains a specific string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
