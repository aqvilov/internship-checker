package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// таймаут на один вызов LLM, согласован отдельно от общего таймаута проверки сайта в checker.go.
const requestTimeout = 15 * time.Second

// Classifier — реализация IClassifier поверх Anthropic API.
type Classifier struct {
	client anthropic.Client
}

func NewClassifier(apiKey string) *Classifier {
	return &Classifier{
		client: anthropic.NewClient(
			option.WithAPIKey(apiKey),
			option.WithRequestTimeout(requestTimeout),
		),
	}
}

// структура ответа модели, задаётся через output_config.format ниже.
type classificationResult struct {
	IsOpen bool   `json:"is_open"`
	Reason string `json:"reason"`
}

var outputSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"is_open": map[string]any{
			"type":        "boolean",
			"description": "true, если фрагмент подтверждает, что набор на стажировку сейчас открыт",
		},
		"reason": map[string]any{
			"type":        "string",
			"description": "краткое обоснование решения на русском языке",
		},
	},
	"required":             []string{"is_open", "reason"},
	"additionalProperties": false,
}

func (c *Classifier) IsInternshipOpen(ctx context.Context, textSnippet, keyword string) (bool, string, error) {
	prompt := fmt.Sprintf(
		"Ниже - фрагмент текста веб-страницы компании, найденный по ключевому слову %q.\n"+
			"Определи, действительно ли фрагмент означает, что набор на стажировку СЕЙЧАС ОТКРЫТ "+
			"(а не закрыт, ещё не начался, проходил в прошлом, или ключевое слово встречается в другом смысле).\n\n"+
			"Фрагмент:\n%s",
		keyword, textSnippet,
	)

	resp, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 300,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		OutputConfig: anthropic.OutputConfigParam{
			Format: anthropic.JSONOutputFormatParam{
				Schema: outputSchema,
			},
		},
	})
	if err != nil {
		return false, "", fmt.Errorf("вызов Anthropic API: %w", err)
	}

	if resp.StopReason == anthropic.StopReasonRefusal {
		return false, "", fmt.Errorf("модель отказалась классифицировать фрагмент")
	}
	if len(resp.Content) == 0 {
		return false, "", fmt.Errorf("пустой ответ от модели")
	}

	var result classificationResult
	if err := json.Unmarshal([]byte(resp.Content[0].Text), &result); err != nil {
		return false, "", fmt.Errorf("разбор ответа модели: %w", err)
	}

	return result.IsOpen, result.Reason, nil
}
