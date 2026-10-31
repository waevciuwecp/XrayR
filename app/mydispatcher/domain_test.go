package mydispatcher

import "testing"

type testMatcher struct{}

func (testMatcher) MatchAny(domain string) bool { return domain == "example.com" }

func TestExcludedDomain(t *testing.T) {
	for _, exclusions := range []any{[]string{"example.com"}, testMatcher{}} {
		if !excludedDomain(exclusions, "example.com") {
			t.Fatal("excluded domain was overridden")
		}
		if excludedDomain(exclusions, "other.com") {
			t.Fatal("unrelated domain was excluded")
		}
	}
	if excludedDomain(nil, "example.com") {
		t.Fatal("nil exclusions matched")
	}
}
