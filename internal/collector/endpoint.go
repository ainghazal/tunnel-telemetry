package collector

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
)

type endpoint struct {
	Proto string
	Host  string
	Port  uint
}

func parseEndpointURI(uri string) (*endpoint, error) {
	e := &endpoint{}

	// Parse the URI
	u, err := url.Parse(uri)
	if err != nil {
		fmt.Println("Error parsing URI:", err)
		return e, err
	}

	// Extract protocol
	e.Proto = u.Scheme

	// Extract host and port
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return e, err
	}
	e.Host = host
	p, err := strconv.Atoi(port)
	if err != nil {
		return e, err
	}
	e.Port = uint(p)
	return e, nil
}
