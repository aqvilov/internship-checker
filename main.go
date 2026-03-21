package main

import (
	"fmt"
	"internship/bot"
	"internship/checker"
	"log"
	"os"
	"time"
)

func main() {
	sites := []checker.Site{
		{Name: "Авито", URL: "https://start.avito.ru/", Keyword: "набор открыт"},
	}

	b := bot.New()
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
			if found {
				b.NotifyAll(fmt.Sprintf("Стажировка у %s открылась: %s\n", site.Name, site.URL))
			}
		}
	}
}
