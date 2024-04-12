package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidMeasurement marks a measurement that does not conform to the minimal sanity checks.
	ErrInvalidMeasurement = errors.New("bad measurement")

	// AllowedClockSkewSeconds is how much the report timestamp can be in the future with regards to server.
	AllowedClockSkewSeconds = 60 * 5

	// AllowedLimitForOldReportsInDays makes server to reject reports with a timestamp older than this.
	AllowedLimitForOldReportsInDays = 7
)

// Failure encapsulates an error reported by clients.
type Failure struct {
	Op    string `json:"op"`
	Error string `json:"error"`
}

// Measurement is a single measurement reported by clients.
type Measurement struct {
	Type         string     `json:"report-type"`
	UUID         string     `json:"uuid,omitempty"`
	OOID         string     `json:"ooni-measurement-id,omitempty"`
	Time         *time.Time `json:"t"`
	Agent        string     `json:"agent,omitempty"`
	Endpoint     string     `json:"endpoint"`
	EndpointAddr string     `json:"endpoint_addr,omitempty"`
	EndpointPort int        `json:"endpoint_port,omitempty"`
	EndpointASN  uint       `json:"endpoint_asn,omitempty"`
	EndpointCC   string     `json:"endpoint_cc,omitempty"`
	Protocol     string     `json:"proto,omitempty"`
	Config       any        `json:"config,omitempty"`
	ClientASN    uint       `json:"client_asn"`
	ClientCC     string     `json:"client_cc"`
	Failure      *Failure   `json:"failure"`
	SamplingRate float32    `json:"sampling_rate"`
}

func NewMeasurement() *Measurement {
	return &Measurement{
		Type:         "",
		UUID:         "",
		OOID:         "",
		Time:         &time.Time{},
		Agent:        "",
		Endpoint:     "",
		EndpointAddr: "",
		EndpointPort: 0,
		EndpointASN:  0,
		Protocol:     "",
		Config:       nil,
		ClientASN:    0,
		ClientCC:     "",
		Failure:      nil,
		SamplingRate: 1,
	}
}

// Validate returns an error if the measurement does not pass sanity checks.
// TODO(ain): pass time object for tests
func (m *Measurement) Validate() error {
	if m.Type == "" {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "type cannot be empty")
	}
	if m.Type != "tunnel-telemetry" {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "type should be 'tunnel-telemetry'")
	}
	if m.Time == nil {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "measurement must send time (t)")
	}
	if m.Time.UTC().After(time.Now().Add(time.Duration(AllowedClockSkewSeconds) * time.Second)) {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "illegal time in the future (t)")
	}
	if m.Time.UTC().Before(time.Now().Add(time.Duration(AllowedLimitForOldReportsInDays) * time.Hour * -24)) {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "invalid time, report too old (t)")
	}
	if m.Endpoint == "" {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "endpoint cannot be empty")
	}
	return nil
}

// PreSave updates any needed fields before saving in the database. It returns an error
// if any of the optional fields are set to improper values.
func (m *Measurement) PreSave() error {
	if m.UUID == "" {
		m.UUID = uuid.New().String()
	}
	// TODO(ain): scrub endpoint IP if configured so.
	return nil
}
