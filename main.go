package main

import (
	"fmt"
	"internship/bot"
	"internship/checker"
	"internship/health"
	"internship/storage"
	"log"
	"os"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	sites := []checker.Site{
		{Name: "Авито", URL: "https://start.avito.ru/", Keyword: "набор открыт"},
		{Name: "Т-Банк", URL: "https://education.tbank.ru/start/go/", Keyword: "подать заявку"},
		{Name: "Kasperskiy", URL: "https://careers.kaspersky.ru/stack/GO", Keyword: "developer go"}, // тут пока простая вака
	}

	connStr := os.Getenv("DATABASE_URL")

	db, err := storage.New(connStr)
	if err != nil {
		log.Fatalf("ошибка подключения к БД: %v", err)
	}
	log.Println("БД подключена")

	// проверяем /health
	go health.StartServer(db.DB())
	//http://localhost:8080/health

	b := bot.New(db, sites) // БЫЛО ТАК b := bot.New(db)
	go b.Start(os.Getenv("TG_TOKEN"))
	chromedp.WaitVisible(".some-specific-selector", chromedp.ByQuery)

	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		var wg sync.WaitGroup
		for _, site := range sites {
			wg.Add(1)
			go func(s checker.Site) {
				defer wg.Done()
				log.Printf("проверяю %s...", s.Name)
				found, err := checker.CheckSite(s.URL, s.Keyword)
				if err != nil {
					log.Println("ошибка:", err)
					return
				}
				if found {
					b.NotifyAll(s.Name, fmt.Sprintf("Стажировка у %s открылась!\nСсылка: %s", s.Name, s.URL))
				}
			}(site)
		}
		wg.Wait()
	}
}
