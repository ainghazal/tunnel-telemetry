package collector

import (
	"fmt"

	"github.com/ainghazal/tunnel-telemetry/internal/config"
	"github.com/ainghazal/tunnel-telemetry/internal/model"
	"github.com/ainghazal/tunnel-telemetry/internal/oonirelay"
	"github.com/ooni/probe-engine/pkg/geoipx"
)

// FileSystemCollector is a simplistic implementation of a collector
// that stores reports in the filesystem.
type FileSystemCollector struct {
	config *config.Config
}

// NewFileSystemCollector creates a new filesystem collector.
func NewFileSystemCollector(cfg *config.Config) *FileSystemCollector {
	return &FileSystemCollector{config: cfg}
}

func (fsc *FileSystemCollector) Geolocate(m *model.Measurement, ip string) error {
	if m.ClientASN != "" && m.ClientCC != "" {
		// the client already filled ASN and CC, so we don't attempt to override it.
		return nil
	}

	asnLookup := mmdbLookupper{}
	if asn, _, err := asnLookup.LookupASN(ip); err == nil {
		m.ClientASN = fmt.Sprintf("AS%d", asn)
	}
	if cc, err := asnLookup.LookupCC(ip); err == nil {
		m.ClientCC = cc
	}

	endpoint, err := parseEndpointURI(m.Endpoint)
	if err != nil {
		return err
	}

	m.Protocol = endpoint.Proto

	if endpoint.Host != "" {

		m.EndpointPort = int(endpoint.Port)

		if fsc.config.AllowPublicEndpoint {
			// we only want to expose the endpoint address if explicitely configured to do so.
			m.EndpointAddr = endpoint.Host
		}
		if asn, _, err := asnLookup.LookupASN(endpoint.Host); err == nil {
			m.EndpointASN = fmt.Sprintf("AS%d", asn)
		}
		if cc, err := asnLookup.LookupCC(endpoint.Host); err == nil {
			m.EndpointCC = cc
		}
	}

	return nil
}

// Save implements [model.Collector]
func (fsc *FileSystemCollector) Save(m *model.Measurement) bool {
	err := m.PreSave(fsc.config)
	return err == nil
}

func (fsc *FileSystemCollector) Submit(mm []*model.Measurement) bool {
	err := oonirelay.SubmitMeasurement(mm[0])
	return err == nil
}

// FileSystemCollector implements [model.GeolocatingCollector]
var _ model.GeolocatingCollector = &FileSystemCollector{}

type mmdbLookupper struct{}

func (mmdbLookupper) LookupASN(ip string) (uint, string, error) {
	return geoipx.LookupASN(ip)
}

func (mmdbLookupper) LookupCC(ip string) (string, error) {
	return geoipx.LookupCC(ip)
}
