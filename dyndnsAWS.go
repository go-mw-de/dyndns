package dyndns

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	log "github.com/sirupsen/logrus"
)

// Errors for specific validation failures.
var (
	ErrInvalidRecordType = errors.New("invalid or forbidden record type")
)

// recordTypeConfig defines permissions for DNS record types.
type recordTypeConfig struct {
	Allow      bool
	ForceAllow bool
}

// recordTypesConfig defines permissions for each DNS record type.
var recordTypesConfig = map[string]recordTypeConfig{
	"A":     {Allow: true, ForceAllow: false},
	"AAAA":  {Allow: true, ForceAllow: false},
	"TXT":   {Allow: true, ForceAllow: false},
	"CNAME": {Allow: true, ForceAllow: false},
	"MX":    {Allow: true, ForceAllow: false},
	"PTR":   {Allow: true, ForceAllow: false},
	"SRV":   {Allow: true, ForceAllow: false},
	"SPF":   {Allow: true, ForceAllow: false},
	"CAA":   {Allow: true, ForceAllow: false},
	"NS":    {Allow: false, ForceAllow: true},
	"SOA":   {Allow: false, ForceAllow: false},
}

// DomainAWS implements the Domain interface for AWS.
type DomainAWS struct {
	Name         string
	ZoneID       string
	AccessKey    string
	AccessSecret string
	TTL          int64
	client       *route53.Route53
}

// recordSetAWS defines the structure for an AWS DNS record.
type recordSetAWS struct {
	names        []string
	value        string
	rsType       string
	ttl          int64
	hostedZoneID string
}

// initClient initializes the AWS Route 53 client.
func (d *DomainAWS) initClient() error {
	if d.client != nil {
		return nil
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(d.AccessKey, d.AccessSecret, ""),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}
	d.client = route53.New(sess)
	return nil
}

// NewAWS creates a new DomainAWS instance.
func NewAWS(name, zoneID, accessKey, accessSecret string, ttl int64) Domain {
	return &DomainAWS{
		Name:         name,
		ZoneID:       zoneID,
		AccessKey:    accessKey,
		AccessSecret: accessSecret,
		TTL:          ttl,
	}
}

// Add adds a new DNS record by calling Update.
func (d DomainAWS) Add(h string, ip string) error {
	log.Infof("Adding DNS record: hostname=%s, ip=%s, zoneID=%s, domain=%s", h, ip, d.ZoneID, d.Name)
	return d.Update(h, ip)
}

// List retrieves all records for a hostname.
func (d DomainAWS) List(h string) ([]Record, error) {
	if err := d.initClient(); err != nil {
		return nil, err
	}

	if err := validateHostname(h); err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(d.ZoneID),
		StartRecordName: aws.String(h),
	}

	var records []Record
	err := d.client.ListResourceRecordSetsPages(input, func(page *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
		for _, rs := range page.ResourceRecordSets {
			if aws.StringValue(rs.Name) != h {
				continue
			}
			var values []string
			for _, rr := range rs.ResourceRecords {
				values = append(values, aws.StringValue(rr.Value))
			}
			records = append(records, Record{
				Name:   aws.StringValue(rs.Name),
				Type:   aws.StringValue(rs.Type),
				TTL:    aws.Int64Value(rs.TTL),
				Values: values,
			})
		}
		return !lastPage
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	log.Infof("Listed %d records for hostname=%s, zoneID=%s", len(records), h, d.ZoneID)
	return records, nil
}

// Delete deletes a specific DNS record for a hostname and IP.
func (d DomainAWS) Delete(h string, ip string, types []string) error {
	if err := d.initClient(); err != nil {
		return err
	}

	if err := validateHostname(h); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	// Validate types
	for _, t := range types {
		config, exists := recordTypesConfig[t]
		if !exists || (!config.Allow && !config.ForceAllow) {
			return fmt.Errorf("%w: %s", ErrInvalidRecordType, t)
		}
	}

	rs, err := rsType(ip)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	// Ensure rs is in types, if types is provided
	if len(types) > 0 && !contains(types, rs) && !contains(types, "TXT") {
		return fmt.Errorf("%w: IP-based record type %s not in requested types %v", ErrInvalidRecordType, rs, types)
	}

	records, err := d.List(h)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	var targetRecord *Record
	for _, record := range records {
		if (record.Type == rs || (contains(types, record.Type) && record.Type == "TXT")) && contains(record.Values, ip) {
			targetRecord = &record
			break
		}
	}

	if targetRecord == nil {
		log.Infof("No matching record found for hostname=%s, ip=%s, types=%v, zoneID=%s", h, ip, types, d.ZoneID)
		return nil
	}

	changeBatch := &route53.ChangeBatch{
		Changes: []*route53.Change{
			{
				Action: aws.String("DELETE"),
				ResourceRecordSet: &route53.ResourceRecordSet{
					Name: aws.String(targetRecord.Name),
					Type: aws.String(targetRecord.Type),
					TTL:  aws.Int64(targetRecord.TTL),
					ResourceRecords: func() []*route53.ResourceRecord {
						var rrs []*route53.ResourceRecord
						for _, v := range targetRecord.Values {
							rrs = append(rrs, &route53.ResourceRecord{Value: aws.String(v)})
						}
						return rrs
					}(),
				},
			},
		},
	}

	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(d.ZoneID),
		ChangeBatch:  changeBatch,
	}

	log.Infof("Deleting record: hostname=%s, type=%s, ttl=%d, values=%v, zoneID=%s",
		targetRecord.Name, targetRecord.Type, targetRecord.TTL, targetRecord.Values, d.ZoneID)

	_, err = d.client.ChangeResourceRecordSets(input)
	if err != nil {
		log.Errorf("Failed to delete record: hostname=%s, type=%s, values=%v: %v",
			targetRecord.Name, targetRecord.Type, targetRecord.Values, err)
		return fmt.Errorf("failed to delete record %s (type=%s): %w", targetRecord.Name, targetRecord.Type, err)
	}

	log.Infof("Successfully deleted record for hostname=%s, type=%s, zoneID=%s", h, targetRecord.Type, d.ZoneID)
	return nil
}

// DeleteRecords deletes DNS records for a hostname based on specified types.
func (d DomainAWS) DeleteRecords(h string, types []string, force bool) error {
	if err := d.initClient(); err != nil {
		return err
	}

	if err := validateHostname(h); err != nil {
		return fmt.Errorf("failed to delete records: %w", err)
	}

	// Validate requested types
	for _, t := range types {
		config, exists := recordTypesConfig[t]
		if !exists || (!config.Allow && !config.ForceAllow) {
			return fmt.Errorf("%w: %s", ErrInvalidRecordType, t)
		}
		if config.ForceAllow && !force {
			return fmt.Errorf("%w: %s requires force=true", ErrInvalidRecordType, t)
		}
	}

	records, err := d.List(h)
	if err != nil {
		return fmt.Errorf("failed to delete records: %w", err)
	}

	if len(records) == 0 {
		log.Infof("No records found to delete for hostname=%s, types=%v, zoneID=%s", h, types, d.ZoneID)
		return nil
	}

	for _, record := range records {
		config, exists := recordTypesConfig[record.Type]
		if !exists || (!config.Allow && !config.ForceAllow) {
			log.Warnf("Skipping deletion of forbidden record type: hostname=%s, type=%s", h, record.Type)
			continue
		}
		if config.ForceAllow && !force {
			log.Warnf("Skipping deletion of record requiring force: hostname=%s, type=%s", h, record.Type)
			continue
		}
		if !contains(types, record.Type) {
			continue
		}

		changeBatch := &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("DELETE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(record.Name),
						Type: aws.String(record.Type),
						TTL:  aws.Int64(record.TTL),
						ResourceRecords: func() []*route53.ResourceRecord {
							var rrs []*route53.ResourceRecord
							for _, v := range record.Values {
								rrs = append(rrs, &route53.ResourceRecord{Value: aws.String(v)})
							}
							return rrs
						}(),
					},
				},
			},
		}

		input := &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(d.ZoneID),
			ChangeBatch:  changeBatch,
		}

		log.Infof("Deleting record: hostname=%s, type=%s, ttl=%d, values=%v, zoneID=%s",
			record.Name, record.Type, record.TTL, record.Values, d.ZoneID)

		_, err := d.client.ChangeResourceRecordSets(input)
		if err != nil {
			log.Errorf("Failed to delete record: hostname=%s, type=%s, values=%v: %v",
				record.Name, record.Type, record.Values, err)
			return fmt.Errorf("failed to delete record %s (type=%s): %w", record.Name, record.Type, err)
		}
	}

	log.Infof("Successfully deleted records for hostname=%s, types=%v, zoneID=%s", h, types, d.ZoneID)
	return nil
}

// Update updates a DNS record for the given hostname and IP.
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
	if err := d.initClient(); err != nil {
		return fmt.Errorf("%T.Update(): %v", d, err)
	}
	log.Infof("%T.Update(): recSet: %#v", d, recSet)
	// Update record
	if _, err := recSet.upsert(d.client); err != nil {
		return fmt.Errorf("%T.Update().recSet.upsert failed: %v", d, err)
	}
	return nil
}

// service creates a new Route 53 client.
func service(akey string, asec string) (*route53.Route53, error) {
	creds := credentials.NewStaticCredentials(akey, asec, "")
	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("(service): %v", err)
	}
	return route53.New(sess, &aws.Config{Credentials: creds}), nil
}

// upsert performs an UPSERT operation for the record set.
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
		ChangeBatch:  &route53.ChangeBatch{Changes: changes},
		HostedZoneId: aws.String(rs.hostedZoneID),
	}
	resp, err := svc.ChangeResourceRecordSets(params)
	if err != nil {
		return nil, fmt.Errorf("(*recordSet).upsert: %v", err)
	}
	return resp, nil
}

// validate checks the record set for validity.
func (rs *recordSetAWS) validate() error {
	for i, name := range rs.names {
		if name == "" {
			return fmt.Errorf("missing record set name at index %d", i)
		}
	}
	if rs.rsType == "" {
		return fmt.Errorf("missing record set type")
	}
	if rs.ttl < 1 {
		return fmt.Errorf("invalid record set TTL: %d", rs.ttl)
	}
	if rs.hostedZoneID == "" {
		return fmt.Errorf("missing hosted zone id")
	}
	return nil
}
