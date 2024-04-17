// Package config contains configuration options for the collector.
package config

// Config allows to customize the server's behavior.
type Config struct {
	// DebugGeolocation configures the insecure defaults for echo server,
	// they allow to spoof the RealIP from the headers.
	DebugGeolocation bool

	// AllowPublicEndpoint keeps the IP of the passed endpoint in the stored reports.
	// When it's set to false (the default) the providers can avoid exposing the IP of
	// the endpoint. For now, we'll be just storing the Port and the ASN of the target endpoint.
	AllowPublicEndpoint bool
}

func NewConfig() *Config {
	return &Config{
		DebugGeolocation: false,
	}
}
