package registry

var (
	registries Collection
)

// Features holds the features a provider supports
type Features []string

// Collection holds a slice of Registry types
type Collection []*Registry

// InitRegistry function for setting connection details of a provider
type InitRegistry func(host, user, pass string) (interface{}, error)

// Registry holds the info about a provider
type Registry struct {
	Provider string
	Protocol string
	InitFn   InitRegistry
	Features Features
}

// Include does the actual work of filtering for specific features
func (rf Features) Include(features ...string) bool {
	if len(features) > len(rf) {
		return false
	}
	fKeys := make(map[string]bool)
	for _, v := range rf {
		fKeys[v] = true
	}
	for _, f := range features {
		if _, ok := fKeys[f]; !ok {
			return false
		}
	}
	return true
}

// Supports does the actual work of filtering for specific features
func (rc Collection) Supports(features ...string) Collection {
	supportedRegistries := make(Collection, 0)
	for _, reg := range rc {
		if reg.Features.Include(features...) {
			supportedRegistries = append(supportedRegistries, reg)
		}
	}
	return supportedRegistries
}

// Using does the actual work of filtering for a specific protocol type
func (rc Collection) Using(proto string) Collection {
	supportedRegistries := make(Collection, 0)
	for _, reg := range rc {
		if reg.Protocol == proto {
			supportedRegistries = append(supportedRegistries, reg)
		}
	}
	return supportedRegistries
}

// For does the actual work of filtering for a specific provider name
func (rc Collection) For(provider string) Collection {
	supportedRegistries := make(Collection, 0)
	for _, reg := range rc {
		if reg.Provider == provider {
			supportedRegistries = append(supportedRegistries, reg)
		}
	}
	return supportedRegistries
}

// All returns all the providers in the registry collection
func All() Collection {
	return registries
}

// Supports will filter the registry collection for providers that support a specific
// implemented feature
func Supports(features ...string) Collection {
	return All().Supports(features...)
}

// Using will filter the registry collection for providers using a specific protocol
func Using(proto string) Collection {
	return All().Using(proto)
}

// For will filter the registry collection for the name of a specific type of provider
func For(provider string) Collection {
	return All().For(provider)
}

// Register will add a provider with details to the main registryCollection
func Register(provider, protocol string, initfn InitRegistry, features []string) {
	regFeatures := make(Features, len(features))
	for i, v := range features {
		regFeatures[i] = v
	}

	registries = append(registries, &Registry{
		Provider: provider,
		Protocol: protocol,
		InitFn:   initfn,
		Features: regFeatures,
	})
}
