package dns

type Zone struct {
	ID         string
	Domain     string
	Namespace  string
	NSPrimary  string
	Hostmaster string
	Serial     uint32
	Refresh    uint32
	Retry      uint32
	Expire     uint32
	Minttl     uint32
	SOATTl     uint32
	Status     string
}

type Record struct {
	ID      string
	ZoneID  string
	FQDN    string
	Type    uint16
	Content string
	TTL     uint32
}
