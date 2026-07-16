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

// держит один общий аллокатор и переиспользует его для всех проверок вместо запуска нового браузера на каждый запрос.
type Checker struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	sem         chan struct{} // ограничивает число одновременно открытых вкладок
}

func NewChecker(maxConcurrent int) *Checker {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.DefaultExecAllocatorOptions[:]...,
	)
	return &Checker{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		sem:         make(chan struct{}, maxConcurrent),
	}
}

func (c *Checker) Close() {
	c.allocCancel()
}

func (c *Checker) Check(ctx context.Context, s Site) (bool, error) {
	return c.CheckSite(ctx, s.URL, s.Keyword)
}

//открывает новую вкладку в уже запущенном браузере. ctx используется для того, чтобы прервать проверку, если программа завершается
func (c *Checker) CheckSite(ctx context.Context, url string, keyword string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	//семафор который ограничивает число открытых вкладок
	select {
	case c.sem <- struct{}{}:
		defer func() { <-c.sem }()
	case <-ctx.Done():
		return false, ctx.Err()
	}

	//таймаут вешаем на переданный ctx, чтобы graceful shutdown действительно прерывал проверку.
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	//вкладка открывается в общем аллокаторе, но с учётом таймаута/отмены из timeoutCtx.
	tabCtx, cancel2 := chromedp.NewContext(c.allocCtx)
	defer cancel2()

	var body string
	err := runWithContext(timeoutCtx, tabCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Text("body", &body, chromedp.ByQuery),
	)
	if err != nil {
		return false, fmt.Errorf("ошибка chromedp: %w", err)
	}

	return strings.Contains(strings.ToLower(body), strings.ToLower(keyword)), nil
}

//запускает действия chromedp в tabCtx, но прерывается, если истекает deadline/отмена timeoutCtx (например, graceful shutdown или общий таймаут в 30с).
func runWithContext(timeoutCtx, tabCtx context.Context, actions ...chromedp.Action) error {
	done := make(chan error, 1)
	go func() {
		done <- chromedp.Run(tabCtx, actions...)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}
