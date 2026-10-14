package bmc

import "context"

// TelemetryServiceInfo describes the Redfish TelemetryService configuration.
type TelemetryServiceInfo struct {
	// ServiceEnabled reports whether the telemetry service is enabled.
	ServiceEnabled bool
	// MaxReports is the maximum number of metric reports supported.
	MaxReports int
	// MinCollectionInterval is the minimum metric collection interval (ISO 8601).
	MinCollectionInterval string
}

// MetricValue is a single metric reading within a metric report.
type MetricValue struct {
	// MetricID is the metric identifier (Redfish MetricId). XCC often leaves
	// this empty and identifies the reading via MetricProperty instead.
	MetricID string
	// MetricProperty is the Redfish resource property the reading came from
	// (e.g. "/redfish/v1/Chassis/1/Power#/PowerControl/0/PowerMetrics/MaxConsumedWatts").
	MetricProperty string
	// Value is the metric value (as reported, stringified).
	Value string
	// Timestamp is the reading timestamp.
	Timestamp string
}

// MetricReportInfo is a metric report and its values.
type MetricReportInfo struct {
	// ID is the Redfish MetricReport Id.
	ID string
	// Name is the report name.
	Name string
	// Values are the report's metric values.
	Values []MetricValue
}

// TelemetryReader is implemented by providers that can read telemetry-service
// metric reports and definitions.
type TelemetryReader interface {
	// TelemetryService returns the telemetry-service configuration.
	TelemetryService(ctx context.Context) (TelemetryServiceInfo, error)
	// MetricReports lists the metric report Ids.
	MetricReports(ctx context.Context) ([]string, error)
	// MetricReport returns a metric report and its values by Id.
	MetricReport(ctx context.Context, id string) (MetricReportInfo, error)
	// MetricReportDefinitions lists the metric report definition Ids.
	MetricReportDefinitions(ctx context.Context) ([]string, error)
	// MetricDefinitions lists the metric definition Ids.
	MetricDefinitions(ctx context.Context) ([]string, error)
	// SubmitTestMetricReport submits a test metric report.
	SubmitTestMetricReport(ctx context.Context, reportName string) error
}
