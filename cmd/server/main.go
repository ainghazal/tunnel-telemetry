package main

import (
	"fmt"
	"net/http"

	"github.com/ainghazal/tt-collector/internal/collector"
	"github.com/ainghazal/tt-collector/internal/model"
	"github.com/labstack/echo/v4"
)

type response struct {
	OK      bool   `json:"ok"`
	Message string `json:"msg"`
}

func main() {
	// TODO(ain): pass config (viper)
	collector := &collector.FileSystemCollector{}
	fmt.Println(collector)

	e := echo.New()
	// We're explicitely setting the IPExtractor to the direct IP Extractor.
	// This means that I need to add the override ability in the case someone needs
	// to setup the collector behind a proxy.
	// TODO: it's a good idea to setup fallback with a flag so that we can test geolocation.
	// e.IPExtractor = echo.ExtractIPDirect()

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, banner)
	})

	e.POST("/report", handleReportCreate(collector))
	e.Logger.Fatal(e.Start(":8080"))
}

func handleReportCreate(c model.Collector) func(echo.Context) error {
	reportCreate := func(ctx echo.Context) error {
		m := model.NewMeasurement()
		if err := ctx.Bind(m); err != nil {
			return ctx.String(http.StatusBadRequest, "bad request: cannot parse json")
		}
		c.Geolocate(m, ctx.RealIP())
		if err := m.Validate(); err != nil {
			r := &response{OK: false, Message: err.Error()}
			return ctx.JSON(http.StatusBadRequest, r)
		}
		m.PreSave()
		return ctx.JSONPretty(http.StatusCreated, m, "  ")
	}
	return reportCreate
}
