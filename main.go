package main

import (
	"fmt"
	"internship/bot"
	"internship/checker"
	"internship/storage"
	"log"
	"os"
	"time"
)

func main() {
	sites := []checker.Site{
		{Name: "Авито", URL: "https://start.avito.ru/", Keyword: "набор открыт"},
	}

	connStr := os.Getenv("DATABASE_URL")

	db, err := storage.New(connStr)
	if err != nil {
		log.Fatalf("ошибка подключения к БД: %v", err)
	}
	log.Println("БД подключена")

	b := bot.New(db)
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
				b.NotifyAll(site.Name, fmt.Sprintf("Стажировка у %s открылась\n%s", site.Name, site.URL))
			}
		}
	}
}
