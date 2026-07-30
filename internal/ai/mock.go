package ai

import "context"

// MockClassifier — тестовая реализация IClassifier с заранее заданным результатом,
// не делающая реальных сетевых вызовов.
type MockClassifier struct {
	IsOpen bool
	Reason string
	Err    error
}

func (m *MockClassifier) IsInternshipOpen(ctx context.Context, textSnippet, keyword string) (bool, string, error) {
	return m.IsOpen, m.Reason, m.Err
}
