// Package oonirelay implements a relay collector that submits sanitized measurements to the OONI Public Collector.
package oonirelay

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ainghazal/tunnel-telemetry/internal/model"
)

var (
	defaultAPI                       = "https://api.dev.ooni.io"
	explorerBase                     = "https://explorer.ooni.org/m/"
	timeFormat                       = "2006-01-02 15:04:05"
	tunnelTelemetryExperimentName    = "tunneltelemetry"
	tunnelTelemetryExperimentVersion = "0.0.1"
	reporterSoftwareName             = "oott"
	reporterSoftwareVersion          = "0.0.1"
)

type ReportRequest struct {
	DataFormatVersion string `json:"data_format_version"`
	Format            string `json:"format"`
	ProbeASN          string `json:"probe_asn"`
	ProbeCC           string `json:"probe_cc"`
	SoftwareName      string `json:"software_name"`
	SoftwareVersion   string `json:"software_version"`
	TestName          string `json:"test_name"`
	TestStartTime     string `json:"test_start_time"`
	TestVersion       string `json:"test_version"`
}

// TODO: pass the Probe(collector) ASN/CC
func NewReportRequest() *ReportRequest {
	rr := &ReportRequest{
		DataFormatVersion: "0.2.0",
		Format:            "json",
		ProbeASN:          "AS32",
		ProbeCC:           "IT",
		SoftwareName:      reporterSoftwareName,
		SoftwareVersion:   reporterSoftwareVersion,
		TestName:          tunnelTelemetryExperimentName,
		TestStartTime:     "",
		TestVersion:       tunnelTelemetryExperimentVersion,
	}
	return rr
}

func (rr *ReportRequest) JSON() ([]byte, error) {
	return json.Marshal(rr)
}

type reportResponse struct {
	BackendVersion string `json:"backend_version"`
	ReportID       string `json:"report_id"`
}

type measurementBody struct {
	MeasurementStartTime string  `json:"measurement_start_time"`
	ProbeASN             string  `json:"probe_asn"`
	ProbeCC              string  `json:"probe_cc"`
	ProbeNetworkName     string  `json:"probe_network_name"`
	ReportID             string  `json:"report_id"`
	SoftwareName         string  `json:"software_name"`
	SoftwareVersion      string  `json:"software_version"`
	TelemetryReport      any     `json:"telemetry_report"`
	TestName             string  `json:"test_name"`
	TestRuntime          float64 `json:"test_runtime"`
	TestStartTime        string  `json:"test_start_time"`
	TestVersion          string  `json:"test_version"`
}

type OONIMeasurement struct {
	Format  string          `json:"format"`
	Content measurementBody `json:"content"`
}

type ooniMeasurementResponse struct {
	MeasurementID string `json:"measurement_uid"`
}

type ReportSubmitter struct {
	API      string
	ReportID string
	Client   *http.Client
}

func NewReportSubmitter() *ReportSubmitter {
	return &ReportSubmitter{
		API:    defaultAPI,
		Client: &http.Client{},
	}
}

func (rs *ReportSubmitter) doPostJSON(url string, data []byte, jd any) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := rs.Client.Do(req)
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

// Start establishes the report channel, and sets the Report ID after a successful acknowledgement.
func (rs *ReportSubmitter) Start(data []byte) error {
	if rs.ReportID != "" {
		return errors.New("already started")
	}
	url := rs.API + "/report"

	var respData reportResponse
	if err := rs.doPostJSON(url, data, &respData); err != nil {
		return err
	}
	rs.ReportID = respData.ReportID
	return nil
}

// SendMeasurement tries to send the passed OONIMeasurement. Returns the measurement ID and any error.
func (rs *ReportSubmitter) SendMeasurement(m *OONIMeasurement) (string, error) {
	if rs.ReportID == "" {
		return "", errors.New("unknown report")
	}
	url := rs.API + "/report/" + rs.ReportID

	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	var measurementData ooniMeasurementResponse
	if err := rs.doPostJSON(url, []byte(data), &measurementData); err != nil {
		return "", err
	}
	return measurementData.MeasurementID, nil
}

// Close sends the closing request for this report channel.
func (rs *ReportSubmitter) Close() error {
	url := rs.API + "/report/" + rs.ReportID + "/close"
	if err := rs.doPostJSON(url, []byte{}, nil); err != nil {
		return err
	}
	return nil
}

// SubmitMeasurement takes a [model.Measurement] and submits a report to OONI.
func SubmitMeasurement(mm *model.Measurement) error {
	rs := NewReportSubmitter()
	rr := NewReportRequest()
	rr.TestStartTime = time.Now().UTC().Format(timeFormat)
	data, err := rr.JSON()
	if err != nil {
		return err
	}

	if err := rs.Start(data); err != nil {
		return err
	}

	m := &OONIMeasurement{
		Format: "json",
		Content: measurementBody{
			MeasurementStartTime: mm.Time.UTC().Format(timeFormat),
			ProbeASN:             "AS32",
			ProbeCC:              "IT",
			ProbeNetworkName:     "Unknown ISP",
			ReportID:             rs.ReportID,
			SoftwareName:         reporterSoftwareName,
			SoftwareVersion:      reporterSoftwareVersion,
			TelemetryReport:      mm,
			TestName:             tunnelTelemetryExperimentName,
			TestRuntime:          0, // it'd be good to add
			TestStartTime:        mm.Time.UTC().Format(timeFormat),
			TestVersion:          tunnelTelemetryExperimentVersion,
		},
	}
	mmid, err := rs.SendMeasurement(m)
	if err != nil {
		return err
	}
	if err := rs.Close(); err != nil {
		return err
	}

	mm.OOID = mmid
	mm.OOIDLink = explorerBase + mmid
	return nil
}