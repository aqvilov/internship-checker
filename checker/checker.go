package checker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// go get github.com/chromedp/chromedp

type Site struct {
	Name    string
	URL     string
	Keyword string
}

func CheckSite(url string, keyword string) (bool, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var body string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.Text("body", &body, chromedp.ByQuery),
	)
	if err != nil {
		return false, fmt.Errorf("ошибка chromedp: %w", err)
	}

	return strings.Contains(strings.ToLower(body), strings.ToLower(keyword)), nil
}
