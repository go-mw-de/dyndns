package dyndns

//AWS DNS Update part of this Programm is based on https://github.com/agorf/dyndns53

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	log "github.com/sirupsen/logrus"
)

type DomainAWS struct {
	Name         string
	ZoneID       string
	AccessKey    string
	AccessSecret string
	TTL          int64
}

// Dns Provider Section
type recordSetAWS struct {
	names        []string
	value        string // ip
	rsType       string // record set type; "A" or "AAAA
	ttl          int64  // TTL (time to live) in seconds
	hostedZoneID string // "hosted zone id"
}

// TBD
func (d DomainAWS) Add(h string, ip string) error {
	log.Fatalf("%T.Add(): Not yet implemented", d)
	return fmt.Errorf("Not yet implemented")
}

// TBD
func (d DomainAWS) Delete(h string, ip string) error {
	log.Fatalf("%T.Delete(): Not yet implemented", d)
	return fmt.Errorf("Not yet implemented")
}

func (d DomainAWS) Update(h string, ip string) error {
	log.Infof("%T.Update(): hostname: %s, domain.ZoneID %s, domain.Name: %s", d, h, d.ZoneID, d.Name)

	// Validate hostname
	if err := validateHostname(h); err != nil {
		return err
	}

	// Determine record type
	rs, err := rsType(ip)
	if err != nil {
		return err
	}

	// Create Change Record
	recSet := recordSetAWS{
		names:        []string{h},
		rsType:       rs,
		ttl:          d.TTL,
		hostedZoneID: d.ZoneID,
		value:        ip,
	}
	svc, err := service(d.AccessKey, d.AccessSecret)
	if err != nil {
		return fmt.Errorf("%T.Update(): %v", d, err)
	}
	log.Infof("%T.Update(): recSet: %#v", d, recSet)

	// Update record

	if _, err := recSet.upsert(svc); err != nil {
		return fmt.Errorf("%T.Update().recSet.upsert failed: %v", d, err)
	}
	return nil
}

func service(akey string, asec string) (*route53.Route53, error) {
	creds := credentials.NewStaticCredentials(akey, asec, "")
	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("(service): %v", err)
	}
	return route53.New(sess, &aws.Config{Credentials: creds}), nil
}

func (rs *recordSetAWS) upsert(svc *route53.Route53) (*route53.ChangeResourceRecordSetsOutput, error) {

	changes := make([]*route53.Change, len(rs.names))
	for i, name := range rs.names {
		changes[i] = &route53.Change{
			Action: aws.String("UPSERT"),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name: aws.String(name),
				Type: aws.String(rs.rsType),
				TTL:  aws.Int64(rs.ttl),
				ResourceRecords: []*route53.ResourceRecord{
					{
						Value: aws.String(rs.value),
					},
				},
			},
		}
	}
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: changes,
		},
		HostedZoneId: aws.String(rs.hostedZoneID),
	}
	resp, err := svc.ChangeResourceRecordSets(params)
	if err != nil {
		return nil, fmt.Errorf("(*recordSet).upsert: %v", err)
	}
	return resp, nil
}

func (rs *recordSetAWS) validate() error {
	for i, name := range rs.names {
		if name == "" {
			return fmt.Errorf("missing record set name at index %d", i)
		}
	}
	if rs.rsType == "" {
		return fmt.Errorf("missing record set type")
	}
	if rs.rsType != "A" && rs.rsType != "AAAA" {
		return fmt.Errorf("invalid record set type: %s", rs.rsType)
	}
	if rs.ttl < 1 {
		return fmt.Errorf("invalid record set TTL: %d", rs.ttl)
	}
	if rs.hostedZoneID == "" {
		return fmt.Errorf("missing hosted zone id")
	}
	return nil
}
