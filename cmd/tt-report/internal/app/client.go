package app

import (
	"log"

	"github.com/ainghazal/tunnel-telemetry/internal/client"
	"github.com/ainghazal/tunnel-telemetry/pkg/geolocate"
)

func processAndSubmitReport(cfg *config) error {
	if cfg.SkipGeolocation {
		log.Println("Skipping geolocation")
	} else {
		geo, err := geolocate.FindCurrentHostGeolocation()
		if err != nil {
			return err
		}
		log.Println("ASN", geo.ASN)
		log.Println("CC", geo.CC)
	}

	_ = client.Client{}

	// c.Submit()
	return nil
}
