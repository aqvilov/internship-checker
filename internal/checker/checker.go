package checker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Site struct {
	Name    string
	URL     string
	Keyword string
}

// Checker держит один общий allocator (запущенный процесс браузера)
// и переиспользует его для всех проверок вместо запуска нового браузера на каждый запрос.
type Checker struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
}

// при старте программы вызывается один ращ
func NewChecker() *Checker {
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.DefaultExecAllocatorOptions[:]...,
	)
	return &Checker{allocCtx: allocCtx, allocCancel: allocCancel}
}

func (c *Checker) Close() {
	c.allocCancel()
}

func (c *Checker) CheckSite(ctx context.Context, url string, keyword string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	timeoutCtx, cancel := context.WithTimeout(c.allocCtx, 30*time.Second)
	defer cancel()

	tabCtx, cancel2 := chromedp.NewContext(timeoutCtx)
	defer cancel2()

	var body string
	err := chromedp.Run(tabCtx,
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
