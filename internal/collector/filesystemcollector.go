package collector

import (
	"log"

	"github.com/ainghazal/tt-collector/internal/model"
	"github.com/ooni/probe-engine/pkg/geoipx"
)

//
// FileSystemCollector is a simplistic implementation of a collector
// that stores reports in the filesystem.
//

type FileSystemCollector struct{}

// TODO: this is done by binding, not needed.
func (fsc *FileSystemCollector) Parse(s string) (*model.Measurement, error) {
	return &model.Measurement{}, nil
}

func (fsc *FileSystemCollector) Geolocate(m *model.Measurement, ip string) error {
	if m.ClientASN != 0 && m.ClientCC != "" {
		// the client already filled ASN and CC, so we don't attempt to override it.
		return nil
	}

	log.Println("DEBUG: ip", ip)

	asnLookup := mmdbLookupper{}
	if asn, _, err := asnLookup.LookupASN(ip); err == nil {
		m.ClientASN = asn
	}
	if cc, err := asnLookup.LookupCC(ip); err == nil {
		m.ClientCC = cc
	}

	endpoint, err := parseEndpointURI(m.Endpoint)
	if err != nil {
		return err
	}

	if endpoint.Host != "" {
		if asn, _, err := asnLookup.LookupASN(endpoint.Host); err == nil {
			m.EndpointASN = asn
		}
		if cc, err := asnLookup.LookupCC(endpoint.Host); err == nil {
			m.EndpointCC = cc
		}
	}

	return nil
}

func (fsc *FileSystemCollector) Save(m *model.Measurement) bool {
	return false
}

// FileSystemCollector implements model.Collector
var _ model.Collector = &FileSystemCollector{}

type mmdbLookupper struct{}

func (mmdbLookupper) LookupASN(ip string) (uint, string, error) {
	return geoipx.LookupASN(ip)
}

func (mmdbLookupper) LookupCC(ip string) (string, error) {
	return geoipx.LookupCC(ip)
}
