package geolocate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
)

var (
	defaultGeolocationAPI = "https://api.dev.ooni.io/api/v1/geolookup"
)

// FindCurrentHostGeolocation will make a best-effor attempt at discovering the public IP
// of the vantage point where the software is running, and obtain geolocation metadata for it.
// This function currently uses a single endpoint for geolocation (in the OONI API).
func FindCurrentHostGeolocation() (*GeoInfo, error) {
	ip, err := AttemptFetchingPublicIP()
	if err != nil {
		return nil, err
	}

	// TODO: use smart-dialer here.
	geo := NewGeolocator()
	info, err := geo.Geolocate(ip)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// AttemptFetchingPublicIP will attempt to get our public IP by exhausting
// all the available sources; the order is stun > https. It will return
// an error if all the sources are used and we still don't have a result.
func AttemptFetchingPublicIP() (string, error) {
	shuffleServers(stunServers)
	for _, server := range stunServers {
		ip, err := FetchIPFromSTUNCall(server)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		fmt.Printf("IP: %s (via %s)\n", ip, server)
		return ip, nil
	}

	shuffleServers(httpsServers)
	for _, provider := range httpsServers {
		fmt.Printf("%s: ", provider)
		ip, err := FetchIPFromHTTPSAPICall(provider)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		} else {
			return ip, nil
		}
	}
	return "", errors.New("out of ideas")
}

// A Geolocator is able to geolocate IPs, using a specific http.Client.
type Geolocator struct {
	API    string
	Client *http.Client
}

// TODO: add NewGeolocationWithHTTPClient
func NewGeolocator() *Geolocator {
	return &Geolocator{
		API:    defaultGeolocationAPI,
		Client: defaultHTTPClient,
	}
}

// GeoInfo contains the minimal metadata that we need for annotating
// reports.
type GeoInfo struct {
	ASName string `json:"as_name"`
	ASN    int    `json:"asn"`
	CC     string `json:"cc"`
}

type geoLocationFromOONI struct {
	Geolocation map[string]GeoInfo `json:"geolocation"`
	Version     int                `json:"v"`
}

func (g *Geolocator) Geolocate(ip string) (*GeoInfo, error) {
	resp := &geoLocationFromOONI{}
	query := fmt.Sprintf(`{"addresses": ["%s"]}`, ip)
	if err := g.doPostJSON(g.API, []byte(query), resp); err != nil {
		return nil, err
	}
	geoinfo := resp.Geolocation[ip]
	return &geoinfo, nil
}

func (g *Geolocator) doPostJSON(url string, data []byte, jd any) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if jd != nil {
		if err := json.NewDecoder(resp.Body).Decode(jd); err != nil {
			return err
		}
	}
	return nil
}

func shuffleServers(ss []string) {
	rand.Shuffle(len(ss), func(i, j int) {
		ss[i], ss[j] = ss[j], ss[i]
	})
}
