// Package client implements a tunneltelemetry server.
package client

import (
	"fmt"

	"github.com/ainghazal/tunnel-telemetry/internal/model"
)

// A Client is used to produce and submit reports to a collector.
type Client struct {
	// If DoGeolocation is set to false, we will not attempt to geolcate ourselves.
	DoGeolocation bool

	// Collector is the primary collector where to submit reports.
	Collector string

	// ClientASN is the ASN for this client's public IP.
	ClientASN string

	// ClientCC is the country code for this client's public IP.
	ClientCC string
}

func (c *Client) Submit(m *model.Measurement) error {
	fmt.Println("dummy submit")
	return nil
}
