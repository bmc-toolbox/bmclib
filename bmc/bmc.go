package bmc

import (
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Metadata represents details about a bmc method
type Metadata struct {
	// SuccessfulProvider is the name of the provider that successfully executed
	SuccessfulProvider string
	// ProvidersAttempted is a slice of all providers that were attempt to execute
	ProvidersAttempted []string
	// SuccessfulOpenConns is a slice of provider names that were opened successfully
	SuccessfulOpenConns []string
	// SuccessfulCloseConns is a slice of provider names that were closed successfully
	SuccessfulCloseConns []string
	// FailedProviderDetail holds the failed providers error messages for called methods
	FailedProviderDetail map[string]string
}

func newMetadata() Metadata {
	return Metadata{
		FailedProviderDetail: make(map[string]string),
	}
}

func (m *Metadata) RegisterSpanAttributes(host string, span trace.Span) {
	span.SetAttributes(attribute.String("host", host))

	span.SetAttributes(attribute.String("successful-provider", m.SuccessfulProvider))

	span.SetAttributes(
		attribute.String("successful-open-conns", strings.Join(m.SuccessfulOpenConns, ",")),
	)

	span.SetAttributes(
		attribute.String("successful-close-conns", strings.Join(m.SuccessfulCloseConns, ",")),
	)

	span.SetAttributes(
		attribute.String("attempted-providers", strings.Join(m.ProvidersAttempted, ",")),
	)

	for p, e := range m.FailedProviderDetail {
		span.SetAttributes(
			attribute.String("provider-errs-"+p, e),
		)
	}
}
