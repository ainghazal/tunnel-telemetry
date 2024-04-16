// package server contains the server abstractions.
package server

import (
	"net/http"

	"github.com/ainghazal/tunnel-telemetry/internal/model"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// Config allows to customize the server's behavior.
type Config struct {
	// DebugGeolocation configures the insecure defaults for echo server,
	// they allow to spoof the RealIP from the headers.
	DebugGeolocation bool
}

func NewConfig() *Config {
	return &Config{
		DebugGeolocation: false,
	}
}

// Response is the result returned by the server.
type Response struct {
	OK      bool   `json:"ok"`
	Message string `json:"msg"`
}

// NewEchoServer returns a configured Echo server.
func NewEchoServer(c *Config) *echo.Echo {
	e := echo.New()
	// We explicitely set IPExtractor to the direct IP Extractor.
	// I will need to add the override ability in the case someone needs
	// to setup the collector behind a proxy; debug mode should only be used for testing.
	if !c.DebugGeolocation {
		e.IPExtractor = echo.ExtractIPDirect()
	}
	//e.Use(middleware.Logger())
	//e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)
	return e
}

// Handler holds methods to handle the different server endpoints.
type Handler struct {
	Collector model.Collector
}

func NewHandler(c model.Collector) *Handler {
	return &Handler{
		Collector: c,
	}
}

// CreateReport creates a new report from client submission.
func (h *Handler) CreateReport(ctx echo.Context) error {
	m := model.NewMeasurement()
	if err := ctx.Bind(m); err != nil {
		return ctx.String(http.StatusBadRequest, "bad request: cannot parse json")
	}
	h.Collector.Geolocate(m, ctx.RealIP())
	if err := m.Validate(); err != nil {
		r := &Response{OK: false, Message: err.Error()}
		return ctx.JSON(http.StatusBadRequest, r)
	}
	// TODO: collector.Save(m) -> should call m.PreSave()
	m.PreSave()
	return ctx.JSONPretty(http.StatusCreated, m, "  ")
}

func HandleRootDecoy(c echo.Context) error {
	return c.HTML(http.StatusOK, decoyBanner)
}
