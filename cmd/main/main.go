package main

import (
	"context"
	"fmt"
	"internship/internal/bot"
	"internship/internal/checker"
	"internship/internal/health"
	"internship/internal/storage"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// ограничивает число одновременно открытых вкладок Chromium при параллельной проверке сайтов.
const maxConcurrentChecks = 2

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("файл .env не найден, использую переменные окружения системы")
	}

	sites := []checker.Site{
		{Name: "Авито", URL: "https://start.avito.ru/", Keyword: "набор открыт"},
		{Name: "Т-Банк", URL: "https://education.tbank.ru/start/go/", Keyword: "подать заявку"},
		{Name: "Kasperskiy", URL: "https://careers.kaspersky.ru/stack/GO", Keyword: "developer go"},
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("переменная окружения DATABASE_URL не задана")
	}

	tgToken := os.Getenv("TG_TOKEN")
	if tgToken == "" {
		log.Fatal("переменная окружения TG_TOKEN не задана")
	}

	db, err := storage.New(connStr)
	if err != nil {
		log.Fatalf("ошибка подключения к БД: %v", err)
	}
	log.Println("БД подключена")

	go health.StartServer(db.DB())

	b := bot.New(db, sites)
	go b.Start(tgToken)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c := checker.NewChecker(maxConcurrentChecks)
	defer c.Close()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("получен сигнал завершения, останавливаюсь...")
			return
		case <-ticker.C:
			var wg sync.WaitGroup
			for _, site := range sites {
				wg.Add(1)
				go func(s checker.Site) {
					defer wg.Done()
					log.Printf("проверяю %s...", s.Name)
					found, err := c.Check(ctx, s)
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
}
