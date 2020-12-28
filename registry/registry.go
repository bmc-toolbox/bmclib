package registry

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
)

var (
	registries Collection
)

const (
	// FeaturePowerState represents the powerstate functionality
	// an implementation will use these when they have implemented
	// corresponding interface method.
	FeaturePowerState    Feature = "powerstate"
	FeaturePowerSet      Feature = "powerset"
	FeatureUserCreate    Feature = "usercreate"
	FeatureUserDelete    Feature = "userdelete"
	FeatureUserUpdate    Feature = "userupdate"
	FeatureUserRead      Feature = "userread"
	FeatureBmcReset      Feature = "bmcreset"
	FeatureBootDeviceSet Feature = "bootdeviceset"
)

// Features holds the features a provider supports
type Features []Feature

// Feature represents a single feature available
type Feature string

// Collection holds a slice of Registry types
type Collection []*Registry

// IsCompatibleFn a function that determines if the
// implementation is compatible with a given BMC
type IsCompatibleFn func(context.Context) bool

// InitRegistry function for setting connection details of a provider
// The return values are as follows:
// interface{} -> the implementation specific struct
// func(context.Context) bool -> a function that determines if the implementation is compatible with a given BMC
// error -> standard error if the initRegistry function fails
type InitRegistry func(host, port, user, pass string, log logr.Logger) (interface{}, IsCompatibleFn, error)

// Registry holds the info about a provider
type Registry struct {
	Provider          string
	Protocol          string
	InitFn            InitRegistry
	Features          Features
	ProviderInterface interface{}
	IsCompatibleFn    IsCompatibleFn
}

// Include does the actual work of filtering for specific features
func (rf Features) Include(features ...Feature) bool {
	if len(features) > len(rf) {
		return false
	}
	fKeys := make(map[Feature]bool)
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
func (rc Collection) Supports(features ...Feature) Collection {
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

// deduplicate returns a new slice with duplicates values removed.
func deduplicate(s []string) []string {
	if len(s) <= 1 {
		return s
	}
	result := []string{}
	seen := make(map[string]struct{})
	for _, val := range s {
		val := strings.ToLower(val)
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

// PreferProtocol does the actual work of moving preferred protocols to the start of the collection
func (rc Collection) PreferProtocol(protocols ...string) Collection {
	var final Collection
	var leftOver Collection
	tracking := make(map[int]Collection)
	protocols = deduplicate(protocols)
	for _, registry := range rc {
		var movedToTracking bool
		for index, pName := range protocols {
			if strings.EqualFold(registry.Protocol, pName) {
				tracking[index] = append(tracking[index], registry)
				movedToTracking = true
			}
		}
		if !movedToTracking {
			leftOver = append(leftOver, registry)
		}
	}
	for x := 0; x <= len(tracking); x++ {
		final = append(final, tracking[x]...)
	}
	final = append(final, leftOver...)
	return final
}

// All returns all the providers in the registry collection
func All() Collection {
	return registries
}

// Supports will filter the registry collection for providers that support a specific
// implemented feature
func Supports(features ...Feature) Collection {
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

// PreferProtocol will move preferred protocols to the start of the collection
func PreferProtocol(protocols ...string) Collection {
	return All().PreferProtocol(protocols...)
}

// Register will add a provider with details to the main registryCollection
func Register(provider, protocol string, initfn InitRegistry, features []Feature) {
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
