package common

import "github.com/xtls/xray-core/features/stats"

// GetOrRegisterCounter supports both the locked core and its newer atomic API.
func GetOrRegisterCounter(manager stats.Manager, name string) (stats.Counter, error) {
	if atomic, ok := manager.(interface {
		GetOrRegisterCounter(string) (stats.Counter, error)
	}); ok {
		return atomic.GetOrRegisterCounter(name)
	}
	if counter := manager.GetCounter(name); counter != nil {
		return counter, nil
	}
	counter, err := manager.RegisterCounter(name)
	if err != nil {
		// Another dispatcher may have registered it concurrently.
		if existing := manager.GetCounter(name); existing != nil {
			return existing, nil
		}
	}
	return counter, err
}
