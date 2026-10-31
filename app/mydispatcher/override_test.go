package mydispatcher

import (
	"context"
	"testing"

	"github.com/xtls/xray-core/common/geodata"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
)

type domainExclusion struct{ geodata.DomainMatcher }

func (domainExclusion) MatchAny(domain string) bool { return domain == "example.com" }

type ipExclusion struct{ geodata.IPMatcher }

func (ipExclusion) Match(ip net.IP) bool { return ip.String() == "192.0.2.1" }

type overrideResult struct{}

func (overrideResult) Domain() string   { return "EXAMPLE.COM" }
func (overrideResult) Protocol() string { return "tls" }

func TestOverrideExclusions(t *testing.T) {
	destination := net.TCPDestination(net.ParseAddress("192.0.2.1"), 443)
	dispatcher := new(DefaultDispatcher)
	for _, tc := range []struct {
		name     string
		domains  geodata.DomainMatcher
		ips      geodata.IPMatcher
		override bool
	}{
		{name: "no-exclusions", override: true},
		{name: "domain", domains: domainExclusion{}},
		{name: "ip", ips: ipExclusion{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := session.SniffingRequest{OverrideDestinationForProtocol: []string{"tls"}, ExcludeForDomain: tc.domains, ExcludeForIP: tc.ips}
			if got := dispatcher.shouldOverride(context.Background(), overrideResult{}, request, destination); got != tc.override {
				t.Fatalf("override = %v, want %v", got, tc.override)
			}
		})
	}
}
