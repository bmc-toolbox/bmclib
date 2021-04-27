package bmc

import "fmt"

// Provider interface describes details about a provider
type Provider interface {
	// Name of the provider
	Name() string
}

// getProviderName returns the name a provider supplies if they implement the Provider interface
// if not implemented then the contrete type is returned
func getProviderName(provider interface{}) string {
	switch p := provider.(type) {
	case Provider:
		return p.Name()
	}
	return fmt.Sprintf("%T", provider)
}
