package common

import (
	"errors"
	"github.com/xtls/xray-core/features/stats"
	"testing"
)

type counter int64

func (c *counter) Value() int64      { return int64(*c) }
func (c *counter) Set(v int64) int64 { old := *c; *c = counter(v); return int64(old) }
func (c *counter) Add(v int64) int64 { *c += counter(v); return int64(*c) }

type legacyManager struct {
	stats.Manager
	value         stats.Counter
	race          bool
	fail          bool
	registrations int
}

func (m *legacyManager) GetCounter(string) stats.Counter { return m.value }
func (m *legacyManager) RegisterCounter(string) (stats.Counter, error) {
	m.registrations++
	if m.fail {
		return nil, errors.New("registration failed")
	}
	m.value = new(counter)
	if m.race {
		return nil, errors.New("already registered")
	}
	return m.value, nil
}

type atomicManager struct {
	stats.Manager
	called bool
}

func (m *atomicManager) GetOrRegisterCounter(string) (stats.Counter, error) {
	m.called = true
	return new(counter), nil
}

func TestCompatibleCounter(t *testing.T) {
	for _, race := range []bool{false, true} {
		m := &legacyManager{race: race}
		first, err := GetOrRegisterCounter(m, "user")
		if err != nil || first == nil {
			t.Fatalf("registration: %v", err)
		}
		first.Add(7)
		second, err := GetOrRegisterCounter(m, "user")
		if err != nil || second != first || second.Value() != 7 || m.registrations != 1 {
			t.Fatal("counter was not reused")
		}
	}
	if _, err := GetOrRegisterCounter(&legacyManager{fail: true}, "user"); err == nil {
		t.Fatal("registration error lost")
	}
	m := &atomicManager{}
	if c, err := GetOrRegisterCounter(m, "user"); err != nil || c == nil || !m.called {
		t.Fatal("atomic method not used")
	}
}
