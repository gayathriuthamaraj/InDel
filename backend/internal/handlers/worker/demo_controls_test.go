package worker

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildSeededOrderSpecUsesCityAreaPoolForLocalOrders(t *testing.T) {
	spec := buildSeededOrderSpec(ZonePair{
		FromCity:  "Hyderabad",
		ToCity:    "Hyderabad",
		FromState: "Telangana",
		ToState:   "Telangana",
	}, 0)

	validAreas := map[string]bool{}
	for _, area := range cityAreaPools["hyderabad"] {
		validAreas[area] = true
	}

	if !validAreas[spec.PickupArea] {
		t.Fatalf("pickup_area = %q, want a Hyderabad area", spec.PickupArea)
	}
	if !validAreas[spec.DropArea] {
		t.Fatalf("drop_area = %q, want a Hyderabad area", spec.DropArea)
	}
	if strings.EqualFold(spec.PickupArea, spec.DropArea) {
		t.Fatalf("pickup_area and drop_area should differ, got %q", spec.PickupArea)
	}
}

func TestHasGenericAreaLabel(t *testing.T) {
	tests := []struct {
		area     string
		city     string
		suffix   string
		expected bool
	}{
		{area: "Pickup Hub 1", city: "Hyderabad", suffix: "Market Road", expected: true},
		{area: "Drop Point 2", city: "Hyderabad", suffix: "Residency", expected: true},
		{area: "Pickup Location", city: "Hyderabad", suffix: "Market Road", expected: true},
		{area: "Hyderabad Market Road", city: "Hyderabad", suffix: "Market Road", expected: true},
		{area: "Banjara Hills", city: "Hyderabad", suffix: "Market Road", expected: false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s_%s", tc.city, tc.area), func(t *testing.T) {
			if got := hasGenericAreaLabel(tc.area, tc.city, tc.suffix); got != tc.expected {
				t.Fatalf("hasGenericAreaLabel(%q, %q, %q) = %t, want %t", tc.area, tc.city, tc.suffix, got, tc.expected)
			}
		})
	}
}

func TestFilterSeedPairsForZoneKeepsOnlyMatchingLocalCity(t *testing.T) {
	pairs := []ZonePair{
		{FromCity: "Hyderabad", ToCity: "Hyderabad", FromState: "Telangana", ToState: "Telangana"},
		{FromCity: "Abhaneri", ToCity: "Abhaneri", FromState: "Rajasthan", ToState: "Rajasthan"},
		{FromCity: "Hyderabad", ToCity: "Warangal", FromState: "Telangana", ToState: "Telangana"},
	}

	filtered := filterSeedPairsForZone(pairs, "Hyderabad", "A")
	if len(filtered) != 1 {
		t.Fatalf("filtered len = %d, want 1", len(filtered))
	}
	if filtered[0].FromCity != "Hyderabad" || filtered[0].ToCity != "Hyderabad" {
		t.Fatalf("filtered pair = %s -> %s, want Hyderabad -> Hyderabad", filtered[0].FromCity, filtered[0].ToCity)
	}
}

func TestCanonicalZoneCityParsesDisplayLabel(t *testing.T) {
	got := canonicalZoneCity("Hyderabad - Telangana", "Hyderabad - Telangana")
	if got != "Hyderabad" {
		t.Fatalf("canonicalZoneCity(...) = %q, want Hyderabad", got)
	}

	normalized := normalizeZoneInput("A", "Hyderabad - Telangana")
	if normalized.City != "Hyderabad" {
		t.Fatalf("normalizeZoneInput city = %q, want Hyderabad", normalized.City)
	}
	if normalized.State != "Telangana" {
		t.Fatalf("normalizeZoneInput state = %q, want Telangana", normalized.State)
	}
}
