//AWS DNS Update part of this Programm is based on https://github.com/agorf/dyndns53

package dyndns

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gitlab.com/echtwerner/appengine/collection"
)

type DomainGAE struct {
	Data collection.Domain
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
