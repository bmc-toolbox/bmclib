package bmc

// Provider interface describes details about a provider
type Provider interface {
	// Name of the provider
	Name() string
}
