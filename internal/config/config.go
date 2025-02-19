// Package config contains configuration options for the collector.
package config

// Config allows to customize the server's behavior.
type Config struct {

	// AllowPublicEndpoint keeps the IP of the passed endpoint in the stored reports.
	// When it's set to false (the default) the providers can avoid exposing the IP of
	// the endpoint. For now, we'll be just storing the Port and the ASN of the target endpoint.
	AllowPublicEndpoint bool

	// AutoTLS enables the use of autocert to automatically fetch LE certificates.
	AutoTLS bool

	// AutoTLSCache is the dir to cache LE TLS material.
	AutoTLSCacheDir string

	// CollectorID is an optional ID to enrich the measurements with.
	CollectorID string

	// Debug sets the debug level in the logs.
	Debug bool

	// DebugGeolocation configures the insecure defaults for echo server,
	// they allow to spoof the RealIP from the headers.
	DebugGeolocation bool

	// Hostname is the domain used for AutoTLS.
	Hostname string

	// ListenAddr is the address where the server lsitens.
	ListenAddr string

	// RelayToOONI will relay reports to OONI if set.
	RelayToOONI bool
}

func NewConfig() *Config {
	return &Config{
		AllowPublicEndpoint: false,
		AutoTLS:             false,
		AutoTLSCacheDir:     "",
		CollectorID:         "",
		Debug:               false,
		DebugGeolocation:    false,
		Hostname:            "",
		RelayToOONI:         false,
	}
}
