package checker

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Site struct {
	Name    string
	URL     string
	Keyword string
}

func CheckSite(url string, keyword string) (bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("ошибка отправки get-запроса %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("ошибка чтения тела %w", err)
	}

	if strings.Contains(strings.ToLower(string(body)), strings.ToLower(keyword)) {
		return true, nil
	}
	return false, nil
}
