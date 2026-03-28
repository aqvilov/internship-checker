package main

import (
	"fmt"
	"internship/bot"
	"internship/checker"
	"internship/health"
	"internship/storage"
	"log"
	"os"
	"time"
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
	time.Sleep(2 * time.Second)

	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		for _, site := range sites {
			log.Printf("проверяю %s...", site.Name)
			found, err := checker.CheckSite(site.URL, site.Keyword)
			if err != nil {
				log.Println("ошибка:", err)
				continue
			}
			log.Printf("found: %v", found)
			if found {
				// чет тут придумать веселое
				b.NotifyAll(site.Name, fmt.Sprintf("Стажировка у %s открылась\nСсылка: %s", site.Name, site.URL))
			}
		}
	}
}
