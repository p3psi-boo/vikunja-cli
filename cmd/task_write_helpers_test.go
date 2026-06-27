package cmd

import (
	"testing"

	"github.com/p3psi-boo/vikunja-cli/model"
)

func TestResolveLabelIDExactCaseInsensitive(t *testing.T) {
	labels := []model.Label{
		{ID: 1, Title: "Backend"},
		{ID: 2, Title: "backend-infra"},
	}
	id, err := resolveLabelID("backend", labels)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Fatalf("expected id 1, got %d", id)
	}
}

func TestResolveLabelIDPrefixFallback(t *testing.T) {
	labels := []model.Label{
		{ID: 5, Title: "Frontend"},
	}
	id, err := resolveLabelID("front", labels)
	if err != nil {
		t.Fatalf("prefix should resolve: %v", err)
	}
	if id != 5 {
		t.Fatalf("expected id 5, got %d", id)
	}
}

func TestResolveLabelIDAmbiguous(t *testing.T) {
	labels := []model.Label{
		{ID: 1, Title: "Backend"},
		{ID: 2, Title: "Backend Ops"},
	}
	_, err := resolveLabelID("back", labels)
	if err == nil {
		t.Fatalf("expected ambiguity error, got nil")
	}
}

func TestResolveLabelIDNotFoundHasGuidance(t *testing.T) {
	_, err := resolveLabelID("nope", []model.Label{{ID: 1, Title: "Other"}})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !contains(err.Error(), "vja label ls") {
		t.Fatalf("error should guide the user, got %q", err.Error())
	}
}

func TestMatchScore(t *testing.T) {
	if !matchScore("backend", "backend", exact) {
		t.Fatal("exact match failed")
	}
	if !matchScore("backend", "back", prefix) {
		t.Fatal("prefix match failed")
	}
	if !matchScore("backend", "ckend", substring) {
		t.Fatal("substring match failed")
	}
	if matchScore("backend", "frontend", exact) {
		t.Fatal("non-match should be false")
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
