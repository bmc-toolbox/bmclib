package registry

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestInclude(t *testing.T) {
	testCases := []struct {
		name     string
		features Features
		includes Features
		want     bool
	}{
		{name: "feature is not included 1", features: Features{}, includes: Features{FeaturePowerSet}, want: false},
		{name: "feature is not included 2", features: Features{FeatureUserCreate}, includes: Features{FeaturePowerSet}, want: false},
		{name: "feature included", features: Features{FeaturePowerSet}, includes: Features{FeaturePowerSet}, want: true},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := tc.features.Include(tc.includes...)
			diff := cmp.Diff(tc.want, result)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestSupports(t *testing.T) {
	testCases := []struct {
		name       string
		collection Collection
		supports   Features
		want       Collection
	}{
		{
			name: "no registry supports UserCreate",
			collection: []*Registry{
				&Registry{
					Provider: "ipmitool",
					Protocol: "ipmi",
					Features: []Feature{FeaturePowerSet},
					InitFn:   nil,
				}},
			supports: Features{
				FeatureUserCreate,
			},
			want: []*Registry{},
		},
		{
			name: "one registry supports UserCreate",
			collection: []*Registry{
				&Registry{
					Provider: "ipmitool",
					Protocol: "ipmi",
					Features: []Feature{FeatureUserCreate},
					InitFn:   nil,
				}},
			supports: Features{
				FeatureUserCreate,
			},
			want: []*Registry{&Registry{
				Provider: "ipmitool",
				Protocol: "ipmi",
				InitFn:   nil,
				Features: []Feature{FeatureUserCreate},
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := tc.collection.Supports(tc.supports...)
			diff := cmp.Diff(tc.want, result)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestUsing(t *testing.T) {
	testCases := []struct {
		name       string
		collection Collection
		proto      string
		want       Collection
	}{
		{
			name: "proto is not found",
			collection: []*Registry{
				&Registry{
					Provider: "ipmitool",
					Protocol: "ipmi",
					Features: []Feature{FeaturePowerSet},
					InitFn:   nil,
				}},
			proto: "web",
			want:  []*Registry{},
		},
		{
			name: "proto is found",
			collection: []*Registry{
				&Registry{
					Provider: "ipmitool",
					Protocol: "ipmi",
					Features: []Feature{FeaturePowerSet},
					InitFn:   nil,
				}},
			proto: "ipmi",
			want: []*Registry{
				&Registry{
					Provider: "ipmitool",
					Protocol: "ipmi",
					Features: []Feature{FeaturePowerSet},
					InitFn:   nil,
				}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := tc.collection.Using(tc.proto)
			diff := cmp.Diff(tc.want, result)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestFor(t *testing.T) {
	testCases := []struct {
		name       string
		collection Collection
		provider   string
		want       Collection
	}{
		{
			name: "proto is not found",
			collection: []*Registry{
				&Registry{
					Provider: "ipmitool",
					Protocol: "ipmi",
					Features: []Feature{FeaturePowerSet},
					InitFn:   nil,
				}},
			provider: "dell",
			want:     []*Registry{},
		},
		{
			name: "proto is found",
			collection: []*Registry{
				&Registry{
					Provider: "ipmitool",
					Protocol: "ipmi",
					Features: []Feature{FeaturePowerSet},
					InitFn:   nil,
				}},
			provider: "ipmitool",
			want: []*Registry{
				&Registry{
					Provider: "ipmitool",
					Protocol: "ipmi",
					Features: []Feature{FeaturePowerSet},
					InitFn:   nil,
				}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := tc.collection.For(tc.provider)
			diff := cmp.Diff(tc.want, result)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestAll(t *testing.T) {
	testCases := []struct {
		name         string
		addARegistry bool
		want         Collection
	}{
		{name: "empty collection", want: nil},
		{name: "single collection", addARegistry: true, want: []*Registry{&Registry{
			Provider: "dell",
			Protocol: "web",
			InitFn:   nil,
			Features: []Feature{FeaturePowerSet},
		}}},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// register
			if tc.addARegistry {
				Register("dell", "web", nil, []Feature{FeaturePowerSet})
			}
			result := All()
			diff := cmp.Diff(tc.want, result)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestSupportFn(t *testing.T) {
	testCases := []struct {
		name         string
		addARegistry bool
		features     []Feature
		want         Collection
	}{
		{name: "empty collection", want: []*Registry{}},
		{name: "single collection", features: []Feature{FeatureUserCreate}, addARegistry: true, want: []*Registry{&Registry{
			Provider: "dell",
			Protocol: "web",
			InitFn:   nil,
			Features: []Feature{FeatureUserCreate},
		}}},
	}
	registries = nil
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// register
			if tc.addARegistry {
				Register("dell", "web", nil, []Feature{FeatureUserCreate})
			}
			result := Supports(tc.features...)
			diff := cmp.Diff(tc.want, result)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestUsingFn(t *testing.T) {
	testCases := []struct {
		name         string
		addARegistry bool
		proto        string
		want         Collection
	}{
		{name: "empty collection", want: []*Registry{}},
		{name: "single collection", proto: "web", addARegistry: true, want: []*Registry{&Registry{
			Provider: "dell",
			Protocol: "web",
			InitFn:   nil,
			Features: []Feature{FeatureUserCreate},
		}}},
	}
	registries = nil
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// register
			if tc.addARegistry {
				Register("dell", "web", nil, []Feature{FeatureUserCreate})
			}
			result := Using(tc.proto)
			diff := cmp.Diff(tc.want, result)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestForFn(t *testing.T) {
	testCases := []struct {
		name         string
		addARegistry bool
		provider     string
		want         Collection
	}{
		{name: "empty collection", want: []*Registry{}},
		{name: "single collection", provider: "dell", addARegistry: true, want: []*Registry{&Registry{
			Provider: "dell",
			Protocol: "web",
			InitFn:   nil,
			Features: []Feature{FeatureUserCreate},
		}}},
	}
	registries = nil
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// register
			if tc.addARegistry {
				Register("dell", "web", nil, []Feature{FeatureUserCreate})
			}
			result := For(tc.provider)
			diff := cmp.Diff(tc.want, result)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestPrefer(t *testing.T) {
	unorderedCollection := Collection{
		{
			Provider: "dell",
			Protocol: "web",
			InitFn:   nil,
			Features: []Feature{FeatureUserCreate},
		},
		{
			Provider: "ipmitool",
			Protocol: "ipmi",
			InitFn:   nil,
			Features: []Feature{FeatureUserCreate},
		},
		{
			Provider: "smc",
			Protocol: "web",
			InitFn:   nil,
			Features: []Feature{FeatureUserCreate},
		},
	}
	testCases := []struct {
		name         string
		addARegistry bool
		protocol     []string
		want         Collection
	}{
		{name: "empty collection", want: unorderedCollection},
		{name: "collection", protocol: []string{"web"}, addARegistry: true, want: Collection{
			{
				Provider: "dell",
				Protocol: "web",
				InitFn:   nil,
				Features: []Feature{FeatureUserCreate},
			},
			{
				Provider: "smc",
				Protocol: "web",
				InitFn:   nil,
				Features: []Feature{FeatureUserCreate},
			},
			{
				Provider: "ipmitool",
				Protocol: "ipmi",
				InitFn:   nil,
				Features: []Feature{FeatureUserCreate},
			},
		}},
		{name: "collection with duplicate protocols", protocol: []string{"web", "web"}, addARegistry: true, want: Collection{
			{
				Provider: "dell",
				Protocol: "web",
				InitFn:   nil,
				Features: []Feature{FeatureUserCreate},
			},
			{
				Provider: "smc",
				Protocol: "web",
				InitFn:   nil,
				Features: []Feature{FeatureUserCreate},
			},
			{
				Provider: "ipmitool",
				Protocol: "ipmi",
				InitFn:   nil,
				Features: []Feature{FeatureUserCreate},
			},
		}},
	}
	registries = nil
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// register
			registries = unorderedCollection
			result := PreferProtocol(tc.protocol...)
			diff := cmp.Diff(tc.want, result)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
