package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ainghazal/tunnel-telemetry/internal/collector"
	"github.com/ainghazal/tunnel-telemetry/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// testFileSystemCollectorWithPayload is an utility function to test handlers exercised by the FileSystemCollector
// implementation.
func testFileSystemCollectorWithPayload(endp, payload string) (echo.Context, *server.Handler, *httptest.ResponseRecorder) {
	e := server.NewEchoServer(server.NewConfig())
	req := httptest.NewRequest(http.MethodPost, endp, strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/report")
	h := server.NewHandler(&collector.FileSystemCollector{})
	return ctx, h, rec
}

func TestRootDecoy(t *testing.T) {
	e := server.NewEchoServer(server.NewConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/")
	if assert.NoError(t, server.HandleRootDecoy(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestMinimalHappyReport(t *testing.T) {
	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		`{
			"report-type": "tunnel-telemetry",
			"t": "2024-04-10T00:00:00Z",
			"endpoint": "ss://1.1.1.1:443",
			"config": {"prefix": "asdf"}}
		'`)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
	}
}

func TestUnknownReportTypeFails(t *testing.T) {
	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		`{
			"report-type": "unknown-telemetry",
			"t": "2024-04-10T00:00:00Z",
			"endpoint": "ss://1.1.1.1:443",
			"config": {"prefix": "asdf"}}
		'`)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestReportFailsWithNoTimestamp(t *testing.T) {
	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		`{
			"report-type": "tunnel-telemetry",
			"endpoint": "ss://1.1.1.1:443",
			"config": {"prefix": "asdf"}}
		'`)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}
