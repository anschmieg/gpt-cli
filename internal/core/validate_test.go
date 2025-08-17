package core

import (
	"testing"
)

func TestProviderOptionsValidate(t *testing.T) {
	opts := &ProviderOptions{APIKey: "", BaseURL: "https://api.test"}
	if err := opts.Validate(); err == nil {
		t.Fatalf("expected validation error for empty API key")
	}

	opts.APIKey = "sk-test"
	if err := opts.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	// nil receiver
	var nilOpts *ProviderOptions
	if err := nilOpts.Validate(); err == nil {
		t.Fatalf("expected error for nil provider options")
	}
}
