package mock

import (
	"context"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// --- Settings ---

func (m *MockDatabasePort) GetSetting(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.Settings[key]
	if !ok {
		return "", domain.ErrNotFound
	}
	return val, nil
}

func (m *MockDatabasePort) SetSetting(ctx context.Context, key, val string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, "SetSetting")

	m.Settings[key] = val
	return nil
}

func (m *MockDatabasePort) ListSettings(ctx context.Context) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make(map[string]string)
	for k, v := range m.Settings {
		res[k] = v
	}
	return res, nil
}
