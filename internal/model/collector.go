package model

// Collector receives measurements and stores them for later processing.
type Collector interface {
	// Save stores a valid Measurement in the internal store.
	Save(m *Measurement) bool
}

// Geolocator is able to extract ASN and Country Code (CC) from the Real IP of the client
// submitting the report.
type Geolocator interface {
	// Geolocate adds geolocation metadata based on the real IP of the client submitting the report.
	Geolocate(m *Measurement, ip string) error
}

type GeolocatingCollector interface {
	Collector
	Geolocator
}

// Submitter sends processed reports or aggregates to an upstream collector.
type Submitter interface {
	// Submit sends a collection of measurements to an upstream collector.
	Submit(mm []*Measurement) bool
}
