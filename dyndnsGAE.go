package dyndns

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type DomainGAE struct {
	Name         string
	ZoneID       string
	AccessKey    string
	AccessSecret string
	TTL          int64
}

type recordSetGAE struct {
	names        []string
	value        string // ip
	rsType       string // record set type; "A" or "AAAA
	ttl          int64  // TTL (time to live) in seconds
	hostedZoneID string // "hosted zone id"
}

// TBD
func (d DomainGAE) Add(h string, ip string) error {
	log.Fatalf("%T.Add(): Not yet implemented", d)
	return fmt.Errorf("Not yet implemented")
}

// TBD
func (d DomainGAE) Delete(h string, ip string) error {
	log.Fatalf("%T.Delete(): Not yet implemented", d)
	return fmt.Errorf("Not yet implemented")
}

// TBD
func (d DomainGAE) Update(h string, ip string) error {
	log.Fatalf("%T.Delete(): Not yet implemented", d)
	return fmt.Errorf("Not yet implemented")
}
