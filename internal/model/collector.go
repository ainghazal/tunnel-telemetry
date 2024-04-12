package model

// Collector receives measurements and stores them for later processing.
type Collector interface {
	// Parse produces a valid Measurement after processing a raw measurement string.
	Parse(s string) (*Measurement, error)

	// Geolocate adds geolocation metadata based on the real IP of the client submitting the report.
	Geolocate(m *Measurement, ip string) error

	// Save stores a valid Measurement in the internal store.
	Save(m *Measurement) bool
}

// Submitter sends processed reports or aggregates to an upstream collector.
type Submitter interface {
	// Submit sends a collection of measurements to an upstream collector.
	Submit(mm []*Measurement) bool
}
