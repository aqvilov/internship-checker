package ai

import "context"

// IClassifier подтверждает через LLM, действительно ли фрагмент текста,
// найденный по ключевому слову, означает открытый набор на стажировку —
// а не ложное совпадение вроде "набор не открыт" или "открыт был в прошлом сезоне".
type IClassifier interface {
	IsInternshipOpen(ctx context.Context, textSnippet, keyword string) (isOpen bool, reason string, err error)
}
