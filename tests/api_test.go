package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/ainghazal/tunnel-telemetry/internal/collector"
	"github.com/ainghazal/tunnel-telemetry/internal/config"
	"github.com/ainghazal/tunnel-telemetry/internal/model"
	"github.com/ainghazal/tunnel-telemetry/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var (
	tsFormat = "2006-01-02T15:04:05Z"
)

type mockRequest struct {
	realIP string
}

// testFileSystemCollectorWithPayload is an utility function to test handlers exercised by the FileSystemCollector
// implementation.
func testFileSystemCollectorWithPayload(endp, payload string, cfg *config.Config, mr *mockRequest) (echo.Context, *server.Handler, *httptest.ResponseRecorder) {
	if cfg == nil {
		cfg = config.NewConfig()
	}
	cfg.DebugGeolocation = true

	e := server.NewEchoServer(cfg)

	req := httptest.NewRequest(http.MethodPost, endp, strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderXForwardedFor, mr.realIP)

	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/report")
	h := server.NewHandler(collector.NewFileSystemCollector(cfg))
	return ctx, h, rec
}

func makeTimestampForYesterday() string {
	return time.Now().Add(time.Hour * (-23)).Format(tsFormat)
}

func makeTimestampForOneMonthAgo() string {
	return time.Now().Add(time.Hour * (-24) * 30).Format(tsFormat)
}

func makeTimestampForTomorrow() string {
	return time.Now().Add(time.Hour * 24).Format(tsFormat)
}

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func TestRootDecoy(t *testing.T) {
	e := server.NewEchoServer(config.NewConfig())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/")
	if assert.NoError(t, server.HandleRootDecoy(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func parseMeasurementResponse(b []byte) (*model.Measurement, error) {
	m := &model.Measurement{}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, err
	}
	return m, nil
}

type reportData struct {
	Type      string
	Timestamp string
	Endpoint  string
	Failure   *string
}

func makeReport(rd *reportData) string {
	reportTmpl := `{
	"report-type": "{{ .Type }}",
	"t": "{{ .Timestamp }}",
	"endpoint": "{{ .Endpoint }}",
	"config": {"prefix": "xx"},
	"failure": {{ if .Failure }}{{ .Failure }}{{ else }}null{{ end }}
}`
	tmpl, _ := template.New("report").Parse(reportTmpl)
	var report bytes.Buffer
	if err := tmpl.Execute(&report, rd); err != nil {
		panic(err)
	}
	return report.String()
}

func TestMinimalHappyReport(t *testing.T) {
	report := makeReport(&reportData{
		Type:      "tunnel-telemetry",
		Timestamp: makeTimestampForYesterday(),
		Endpoint:  "ss://1.1.1.1:443",
	})

	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		report,
		nil,
		&mockRequest{},
	)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		if assert.Equal(t, http.StatusCreated, rec.Code) {
			m, err := parseMeasurementResponse(rec.Body.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			assert.True(t, isValidUUID(m.UUID))
			assert.Equal(t, "ss", m.Protocol)
			assert.Equal(t, 443, m.EndpointPort)
			// by default, we scrub the endpoint field, so we don't expect
			// the endpoint address to be public either.
			assert.Equal(t, "", m.EndpointAddr)
			assert.Equal(t, "", m.Endpoint)
		}
	}
}

func TestMinimalHappyReportWithPublicEndpointSetting(t *testing.T) {
	report := makeReport(&reportData{
		Type:      "tunnel-telemetry",
		Timestamp: makeTimestampForYesterday(),
		Endpoint:  "ss://1.1.1.1:443",
	})

	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		report,
		&config.Config{
			AllowPublicEndpoint: true,
		},
		&mockRequest{},
	)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		if assert.Equal(t, http.StatusCreated, rec.Code) {
			m, err := parseMeasurementResponse(rec.Body.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			assert.True(t, isValidUUID(m.UUID))
			assert.Equal(t, "ss", m.Protocol)
			assert.Equal(t, 443, m.EndpointPort)
			// in this test we do allow public endpoint collection, so this collector
			// should not scrub the endpoint IP address / hostname.
			assert.Equal(t, "1.1.1.1", m.EndpointAddr)
			assert.Equal(t, "ss://1.1.1.1:443", m.Endpoint)
		}
	}
}

func TestReportWithFailure(t *testing.T) {
	failure := `{"op": "dns", "error": "dns.cannot_resolve"}`
	report := makeReport(&reportData{
		Type:      "tunnel-telemetry",
		Timestamp: makeTimestampForYesterday(),
		Endpoint:  "ss://1.1.1.1:443",
		Failure:   &failure,
	})
	fmt.Println(report)
	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		report,
		nil,
		&mockRequest{},
	)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		if assert.Equal(t, http.StatusCreated, rec.Code) {
			m, err := parseMeasurementResponse(rec.Body.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, &model.Failure{Op: "dns", Error: "dns.cannot_resolve"}, m.Failure)
		}
	}
}

func TestUnknownReportTypeFails(t *testing.T) {
	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		makeReport(&reportData{
			Type:      "unknown-telemetry",
			Timestamp: makeTimestampForYesterday(),
			Endpoint:  "ss://1.1.1.1:443",
		}),
		nil,
		&mockRequest{},
	)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestReportFailsWithNoTimestamp(t *testing.T) {
	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		makeReport(&reportData{
			Type:      "tunnel-telemetry",
			Timestamp: "",
			Endpoint:  "ss://1.1.1.1:443",
		}),
		nil,
		&mockRequest{},
	)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestReportFailsWithTimestampTooOld(t *testing.T) {
	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		makeReport(&reportData{
			Type:      "tunnel-telemetry",
			Timestamp: makeTimestampForOneMonthAgo(),
			Endpoint:  "ss://1.1.1.1:443",
		}),
		nil,
		&mockRequest{},
	)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestReportFailsWithTimestampInTheFuture(t *testing.T) {
	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		makeReport(&reportData{
			Type:      "tunnel-telemetry",
			Timestamp: makeTimestampForTomorrow(),
			Endpoint:  "ss://1.1.1.1:443",
		}),
		nil,
		&mockRequest{},
	)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestClientGeolocationWithSpoofedHeader(t *testing.T) {
	ctx, hdlr, rec := testFileSystemCollectorWithPayload(
		"/report",
		makeReport(&reportData{
			Type:      "tunnel-telemetry",
			Timestamp: makeTimestampForYesterday(),
			Endpoint:  "ss://1.1.1.1:443",
		}),
		nil,
		&mockRequest{realIP: "2.3.4.5"},
	)
	if assert.NoError(t, hdlr.CreateReport(ctx)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		m, err := parseMeasurementResponse(rec.Body.Bytes())
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, m.ClientCC, "FR")
		assert.Equal(t, m.ClientASN, uint(3215))
		assert.Equal(t, m.EndpointCC, "AU")
		assert.Equal(t, m.EndpointASN, uint(13335))
	}
}
