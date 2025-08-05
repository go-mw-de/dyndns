package dyndns

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// DomainGAE implements the Domain interface for GAE (placeholder).
type DomainGAE struct {
	Name         string
	ZoneID       string
	AccessKey    string
	AccessSecret string
	TTL          int64
}

// recordSetGAE defines the structure for a GAE DNS record.
type recordSetGAE struct {
	names        []string
	value        string
	rsType       string
	ttl          int64
	hostedZoneID string
}

// NewGAE creates a new DomainGAE instance.
func NewGAE(name, zoneID, accessKey, accessSecret string, ttl int64) Domain {
	return &DomainGAE{
		Name:         name,
		ZoneID:       zoneID,
		AccessKey:    accessKey,
		AccessSecret: accessSecret,
		TTL:          ttl,
	}
}

// Add adds a new DNS record (not implemented).
func (d DomainGAE) Add(h string, ip string) error {
	log.Errorf("%T.Add(): Not yet implemented", d)
	return fmt.Errorf("Not yet implemented")
}

// Delete deletes a specific DNS record (not implemented).
func (d DomainGAE) Delete(h string, ip string, types []string) error {
	log.Errorf("%T.Delete(): Not yet implemented", d)
	return fmt.Errorf("Not yet implemented")
}

// Update updates a DNS record (not implemented).
func (d DomainGAE) Update(h string, ip string) error {
	log.Errorf("%T.Update(): Not yet implemented", d)
	return fmt.Errorf("Not yet implemented")
}

// List retrieves all records for a hostname (not implemented).
func (d DomainGAE) List(h string) ([]Record, error) {
	log.Errorf("%T.List(): Not yet implemented", d)
	return nil, fmt.Errorf("Not yet implemented")
}

// DeleteRecords deletes DNS records for a hostname (not implemented).
func (d DomainGAE) DeleteRecords(h string, types []string, force bool) error {
	log.Errorf("%T.DeleteRecords(): Not yet implemented", d)
	return fmt.Errorf("Not yet implemented")
}
