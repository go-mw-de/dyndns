package dyndns

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
