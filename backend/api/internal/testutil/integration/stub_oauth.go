package integration

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/mq/api/internal/port/outbound"
)

// StubGoogle stands in for Google in integration tests: no network, the test decides who "Google" says the
// person is. Grant(code, info) makes the next exchange of that code return info.
type StubGoogle struct {
	mu     sync.Mutex
	grants map[string]outbound.OAuthUserInfo
}

func NewStubGoogle() *StubGoogle { return &StubGoogle{grants: map[string]outbound.OAuthUserInfo{}} }

func (g *StubGoogle) Name() string { return "google" }

func (g *StubGoogle) AuthCodeURL(state string) string {
	return "https://accounts.google.test/o/oauth2/auth?state=" + url.QueryEscape(state)
}

func (g *StubGoogle) Exchange(_ context.Context, code string) (*outbound.OAuthUserInfo, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	info, ok := g.grants[code]
	if !ok {
		return nil, fmt.Errorf("stub google: unknown code %q", code)
	}
	return &info, nil
}

// Grant registers what Google will report for code.
func (g *StubGoogle) Grant(code string, info outbound.OAuthUserInfo) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.grants[code] = info
}
