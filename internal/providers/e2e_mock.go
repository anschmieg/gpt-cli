//go:build e2e
// +build e2e

package providers

import "github.com/anschmieg/gpt-cli/internal/config"

// E2EMockProvider is a request-aware mock used by e2e tests. It routes calls
// through package-level responder functions that tests can set.
type E2EMockProvider struct{ name string }

// E2ECallResponder, if non-nil, handles CallProvider for e2e tests.
var E2ECallResponder func(providerName, prompt string) (string, error)

// E2EStreamResponder, if non-nil, handles StreamProvider for e2e tests.
var E2EStreamResponder func(providerName, prompt string) (<-chan string, <-chan error)

func (m E2EMockProvider) GetName() string { return m.name }

func (m E2EMockProvider) CallProvider(prompt string) (string, error) {
    if E2ECallResponder != nil {
        return E2ECallResponder(m.name, prompt)
    }
    return "", nil
}

func (m E2EMockProvider) StreamProvider(prompt string) (<-chan string, <-chan error) {
    if E2EStreamResponder != nil {
        return E2EStreamResponder(m.name, prompt)
    }
    c := make(chan string); e := make(chan error)
    close(c); close(e)
    return c, e
}

// In e2e builds, override provider creation via the hook to return the mock.
func init() {
    NewProviderHook = func(providerName string, cfg *config.Config) Provider {
        n := providerName
        if n != "openai" && n != "copilot" && n != "gemini" {
            n = "openai"
        }
        return E2EMockProvider{name: n}
    }
}
