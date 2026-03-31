package dns

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/rdata"
	"github.com/hicompute/kloudstack/pkg/daemons/dns/pb"
	"github.com/scylladb/gocqlx/v2"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
)

func (s *dnsServer) Query(ctx context.Context, req *pb.DnsPacket) (*pb.DnsPacket, error) {
	msg := new(dns.Msg)
	msg.Data = req.Msg
	if err := msg.Unpack(); err != nil {
		return nil, err
	}

	resp := new(dns.Msg)
	resp.Response = true
	resp.Authoritative = true
	resp.ID = msg.ID
	resp.Opcode = dns.OpcodeQuery
	resp.RecursionAvailable = false
	resp.Question = msg.Question

	for _, q := range msg.Question {
		fqdn := normalizeFQDN(q.Header().Name)
		qtype := dns.RRToType(q)

		// 🔥 FAST PATH
		records, err := s.lookupActiveRecords(fqdn, qtype)
		if err != nil {
			return nil, err
		}

		if len(records) > 0 {
			resp.Rcode = dns.RcodeSuccess
			s.appendAnswers(resp, records, qtype)
			continue
		}

		// 🔻 MISS → check zone
		zone, err := s.findActiveZone(fqdn)
		if err != nil {
			return nil, err
		}

		if zone == nil {
			resp.Rcode = dns.RcodeNameError // NXDOMAIN
			resp.Ns = append(resp.Ns, s.getDefaultSOA())
			continue
		}

		// NODATA
		resp.Rcode = dns.RcodeSuccess
		resp.Ns = append(resp.Ns, s.getSOARecordForZone(zone))
	}

	if err := resp.Pack(); err != nil {
		return nil, err
	}

	return &pb.DnsPacket{Msg: resp.Data}, nil
}

// Look up records within a specific zone
func (s *dnsServer) lookupActiveRecords(fqdn string, qtype uint16) ([]Record, error) {
	// Step 1: CNAME first (DNS rule)
	cname, _ := s.queryActiveRecords(fqdn, dns.TypeCNAME)
	if len(cname) > 0 {
		return cname, nil
	}

	// Step 2: exact match
	records, _ := s.queryActiveRecords(fqdn, qtype)
	if len(records) > 0 {
		return records, nil
	}

	// Step 3: wildcard fallback
	parts := strings.Split(fqdn, ".")
	for i := 1; i < len(parts); i++ {
		wildcard := "*." + strings.Join(parts[i:], ".")
		records, _ := s.queryActiveRecords(wildcard, qtype)
		if len(records) > 0 {
			return records, nil
		}
	}

	return nil, nil
}

// Query records from database
func (s *dnsServer) queryActiveRecords(fqdn string, qtype uint16) ([]Record, error) {
	query := `SELECT zone_id, fqdn, type, content, ttl FROM active_records WHERE fqdn = :fqdn`
	if qtype == dns.TypeA || qtype == dns.TypeAAAA {
		query += fmt.Sprintf(` AND type IN ( %d, %d, %d )`, dns.TypeA, dns.TypeAAAA, dns.TypeCNAME)
	} else {
		query += ` AND type = :type`
	}
	var iter *gocqlx.Iterx
	var records []Record

	if qtype == dns.TypeA || qtype == dns.TypeAAAA || qtype == dns.TypeCNAME {
		iter = s.db.Query(query, []string{":fqdn"}).BindMap(map[string]any{":fqdn": fqdn}).Iter()
	} else {
		iter = s.db.Query(query, []string{":fqdn", ":type"}).BindMap(map[string]any{":fqdn": fqdn, ":type": qtype}).Iter()
	}
	var r Record

	for iter.Scan(&r.ZoneID, &r.FQDN, &r.Type, &r.Content, &r.TTL) {
		records = append(records, r)
	}

	klog.Infof("found records for %s, %v", fqdn, records)
	return records, iter.Close()
}

func (s *dnsServer) findActiveZone(fqdn string) (*Zone, error) {
	parts := strings.Split(fqdn, ".")

	for i, _ := range parts {
		domain := strings.Join(parts[i:], ".")

		query := `SELECT zone_id FROM active_zones WHERE domain = :domain`

		var zoneID string
		iter := s.db.Query(query, []string{":domain"}).BindMap(map[string]any{
			":domain": domain,
		}).Iter()

		if iter.Scan(&zoneID) {
			iter.Close()
			return s.getZoneByID(zoneID)
		}
		iter.Close()
	}

	return nil, nil
}

// Get SOA record for a zone (from database)
func (s *dnsServer) getSOARecordForZone(zone *Zone) dns.RR {
	return &dns.SOA{
		Hdr: dns.Header{
			Name:  zone.Domain,
			Class: dns.ClassINET,
			TTL:   zone.SOATTl,
		},
		SOA: rdata.SOA{
			Ns:      zone.NSPrimary,
			Mbox:    zone.Hostmaster,
			Serial:  zone.Serial,
			Refresh: zone.Refresh,
			Retry:   zone.Retry,
			Expire:  zone.Expire,
			Minttl:  zone.Minttl,
		},
	}
}

func (s *dnsServer) appendAnswers(resp *dns.Msg, records []Record, qtype uint16) {
	for _, r := range records {
		rr := s.createRRFromRecord(r)
		if rr != nil {
			resp.Answer = append(resp.Answer, rr)
		}

		// CNAME resolution
		if r.Type == dns.TypeCNAME && qtype != dns.TypeCNAME {
			target := normalizeFQDN(r.Content)
			targetRecords, _ := s.lookupActiveRecords(target, qtype)

			for _, tr := range targetRecords {
				rr := s.createRRFromRecord(tr)
				if rr != nil {
					resp.Answer = append(resp.Answer, rr)
				}
			}
		}
	}
}

func (s *dnsServer) createRRFromRecord(record Record) dns.RR {
	switch record.Type {
	case dns.TypeTXT:
		return &dns.TXT{
			Hdr: dns.Header{
				Name:  record.FQDN,
				Class: dns.ClassINET,
				TTL:   record.TTL,
			},
			TXT: rdata.TXT{Txt: []string{record.Content}},
		}
	case dns.TypeA:
		ip, _ := netip.ParseAddr(record.Content)
		return &dns.A{
			Hdr: dns.Header{
				Name:  record.FQDN,
				Class: dns.ClassINET,
				TTL:   record.TTL,
			},
			A: rdata.A{Addr: ip},
		}
	case dns.TypeAAAA:
		ip, _ := netip.ParseAddr(record.Content)
		return &dns.AAAA{
			Hdr: dns.Header{
				Name:  record.FQDN,
				Class: dns.ClassINET,
				TTL:   record.TTL,
			},
			AAAA: rdata.AAAA{Addr: ip},
		}
	case dns.TypeCNAME:
		return &dns.CNAME{
			Hdr: dns.Header{
				Name:  record.FQDN,
				Class: dns.ClassINET,
				TTL:   record.TTL,
			},
			CNAME: rdata.CNAME{Target: record.Content},
		}
	}
	return nil
}

// getDefaultSOA returns a default SOA record for non-existent domains
// This is optional - some DNS servers don't add SOA for NXDOMAIN responses
func (s *dnsServer) getDefaultSOA() dns.RR {
	// Get from config or use hardcoded values
	ns := viper.GetString("DNS_M_NAME")
	hostmaster := viper.GetString("DNS_R_NAME")

	return &dns.SOA{
		Hdr: dns.Header{
			Name:  ".", // Root zone, or you might want to use the TLD
			Class: dns.ClassINET,
			TTL:   3600,
		},
		SOA: rdata.SOA{
			Ns:      ns,
			Mbox:    hostmaster,
			Serial:  uint32(time.Now().Unix()), // Dynamic serial
			Refresh: 3600,
			Retry:   600,
			Expire:  86400,
			Minttl:  300,
		},
	}
}

func normalizeFQDN(fqdn string) string {
	fqdn = strings.ToLower(fqdn)
	if !strings.HasSuffix(fqdn, ".") {
		fqdn += "."
	}
	return fqdn
}

func (s *dnsServer) getZoneByID(zoneID string) (*Zone, error) {
	query := `
	SELECT id, domain, namespace, ns_primary, hostmaster,
	       serial, refresh, retry, expire, minttl, soa_ttl
	FROM zones_by_id
	WHERE id = :id
	`

	var zone Zone

	iter := s.db.Query(query, []string{":id"}).BindMap(map[string]any{
		":id": zoneID,
	}).Iter()

	if !iter.Scan(
		&zone.ID,
		&zone.Domain,
		&zone.Namespace,
		&zone.NSPrimary,
		&zone.Hostmaster,
		&zone.Serial,
		&zone.Refresh,
		&zone.Retry,
		&zone.Expire,
		&zone.Minttl,
		&zone.SOATTl,
	) {
		iter.Close()
		return nil, nil
	}

	return &zone, iter.Close()
}
