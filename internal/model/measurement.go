package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/ainghazal/tunnel-telemetry/internal/config"

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
	OOIDLink     string     `json:"ooni-measurement-link,omitempty"`
	TimeStart    *time.Time `json:"time"`
	DurationMS   int64      `json:"duration_ms,omitempty"`
	TimeReported *time.Time `json:"t_reported,omitempty"`
	TimeRelayed  *time.Time `json:"t_relayed,omitempty"`
	Agent        string     `json:"agent,omitempty"`
	CollectorID  string     `json:"collector_id,omitempty"`
	Endpoint     string     `json:"endpoint,omitempty"`
	EndpointAddr string     `json:"endpoint_addr,omitempty"`
	EndpointPort int        `json:"endpoint_port,omitempty"`
	EndpointASN  string     `json:"endpoint_asn,omitempty"`
	EndpointCC   string     `json:"endpoint_cc,omitempty"`
	Protocol     string     `json:"proto,omitempty"`
	Config       any        `json:"config,omitempty"`
	ClientASN    string     `json:"client_asn"`
	ClientCC     string     `json:"client_cc"`
	Failure      *Failure   `json:"failure,omitempty"`
	SamplingRate float32    `json:"sampling_rate"`
}

func NewMeasurement() *Measurement {
	return &Measurement{
		Type:         "",
		UUID:         "",
		OOID:         "",
		TimeStart:    &time.Time{},
		DurationMS:   0,
		Agent:        "",
		CollectorID:  "",
		Endpoint:     "",
		EndpointAddr: "",
		EndpointPort: 0,
		EndpointASN:  "",
		Protocol:     "",
		Config:       nil,
		ClientASN:    "",
		ClientCC:     "",
		Failure:      nil,
		SamplingRate: 1.0,
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
	if m.TimeStart == nil {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "measurement must send time (t)")
	}
	if m.TimeStart.UTC().After(time.Now().Add(time.Duration(AllowedClockSkewSeconds) * time.Second)) {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "illegal time in the future (t)")
	}
	if m.TimeStart.UTC().Before(time.Now().Add(time.Duration(AllowedLimitForOldReportsInDays) * time.Hour * -24)) {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "invalid time, report too old (t)")
	}
	if m.DurationMS < 0 {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "duration cannot be negative")
	}
	if m.Endpoint == "" {
		return fmt.Errorf("%w: %s", ErrInvalidMeasurement, "endpoint cannot be empty")
	}
	return nil
}

// PreSave updates any needed fields before saving in the database. It returns an error
// if any of the optional fields are set to improper values.
func (m *Measurement) PreSave(cfg *config.Config) error {
	if m.UUID == "" {
		// assign a UUID if the report did not have one.
		m.UUID = uuid.New().String()
	}
	if !cfg.AllowPublicEndpoint {
		// scrub the endpoint IP Address.
		m.Endpoint = ""
	}
	if cfg.CollectorID != "" {
		m.CollectorID = cfg.CollectorID
	}
	return nil
}
