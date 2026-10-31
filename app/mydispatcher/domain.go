package mydispatcher

// The locked core uses a slice; newer cores compile domain matchers.
func excludedDomain(exclusions any, domain string) bool {
	switch value := exclusions.(type) {
	case []string:
		for _, excluded := range value {
			if excluded == domain {
				return true
			}
		}
	case interface{ MatchAny(string) bool }:
		return value.MatchAny(domain)
	}
	return false
}
