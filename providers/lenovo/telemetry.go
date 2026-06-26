package lenovo

import (
	"context"
	"net/url"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

const (
	telemetryServiceURI        = "/redfish/v1/TelemetryService"
	metricReportsURI           = telemetryServiceURI + "/MetricReports"
	metricReportDefinitionsURI = telemetryServiceURI + "/MetricReportDefinitions"
	metricDefinitionsURI       = telemetryServiceURI + "/MetricDefinitions"
	submitTestMetricReportURI  = telemetryServiceURI + "/Actions/TelemetryService.SubmitTestMetricReport"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.TelemetryReader = (*Conn)(nil)

// TelemetryService returns the XCC telemetry-service configuration.
//
// Implements bmc.TelemetryReader.
func (c *Conn) TelemetryService(ctx context.Context) (bmc.TelemetryServiceInfo, error) {
	var doc struct {
		ServiceEnabled        bool   `json:"ServiceEnabled"`
		MaxReports            int    `json:"MaxReports"`
		MinCollectionInterval string `json:"MinCollectionInterval"`
	}
	if err := c.getJSON(telemetryServiceURI, &doc); err != nil {
		return bmc.TelemetryServiceInfo{}, err
	}

	return bmc.TelemetryServiceInfo{
		ServiceEnabled:        doc.ServiceEnabled,
		MaxReports:            doc.MaxReports,
		MinCollectionInterval: doc.MinCollectionInterval,
	}, nil
}

// MetricReports lists the metric report Ids.
//
// Implements bmc.TelemetryReader.
func (c *Conn) MetricReports(ctx context.Context) ([]string, error) {
	return c.memberIDs(metricReportsURI)
}

// MetricReport returns a metric report and its values by Id.
//
// Implements bmc.TelemetryReader.
func (c *Conn) MetricReport(ctx context.Context, id string) (bmc.MetricReportInfo, error) {
	var doc struct {
		ID           string `json:"Id"`
		Name         string `json:"Name"`
		MetricValues []struct {
			MetricID       string `json:"MetricId"`
			MetricProperty string `json:"MetricProperty"`
			MetricValue    string `json:"MetricValue"`
			Timestamp      string `json:"Timestamp"`
		} `json:"MetricValues"`
	}
	reportURL, err := url.JoinPath(metricReportsURI, id)
	if err != nil {
		return bmc.MetricReportInfo{}, err
	}
	if err := c.getJSON(reportURL, &doc); err != nil {
		return bmc.MetricReportInfo{}, err
	}

	report := bmc.MetricReportInfo{ID: doc.ID, Name: doc.Name}
	for _, v := range doc.MetricValues {
		report.Values = append(report.Values, bmc.MetricValue{
			MetricID:       v.MetricID,
			MetricProperty: v.MetricProperty,
			Value:          v.MetricValue,
			Timestamp:      v.Timestamp,
		})
	}

	return report, nil
}

// MetricReportDefinitions lists the metric report definition Ids.
//
// Implements bmc.TelemetryReader.
func (c *Conn) MetricReportDefinitions(ctx context.Context) ([]string, error) {
	return c.memberIDs(metricReportDefinitionsURI)
}

// MetricDefinitions lists the metric definition Ids.
//
// Implements bmc.TelemetryReader.
func (c *Conn) MetricDefinitions(ctx context.Context) ([]string, error) {
	return c.memberIDs(metricDefinitionsURI)
}

// SubmitTestMetricReport submits a test metric report.
//
// Implements bmc.TelemetryReader.
func (c *Conn) SubmitTestMetricReport(ctx context.Context, reportName string) error {
	payload := map[string]any{}
	if reportName != "" {
		payload["MetricReportName"] = reportName
	}

	return checkResponse(c.redfishwrapper.PostWithHeaders(ctx, submitTestMetricReportURI, payload, nil))
}
