package bmc

// Metadata represents details about a bmc method
type Metadata struct {
	// SuccessfulProvider is the name of the provider that successfully executed
	SuccessfulProvider string
	// ProvidersAttempted is a slice of all providers that were attempt to execute
	ProvidersAttempted []string
}
