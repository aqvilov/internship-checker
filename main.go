package main

import (
	"internship/bot"
	"internship/checker"
	"log"
	"time"
)

func main() {
	b := bot.New()

	go b.Start("8638425009:AAFsF0IJ_LzISw5PInnhctoCd7aenzbCWEo")
	time.Sleep(2 * time.Second)

	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		log.Println("проверяю сайт...")
		found, err := checker.CheckSite("https://start.avito.ru/", "набор открыт")
		if err != nil {
			log.Println("ошибка:", err)
			continue
		}
		log.Println("found:", found)
		if found {
			b.NotifyAll("Стажировка открылась!")
		}
	}

}
