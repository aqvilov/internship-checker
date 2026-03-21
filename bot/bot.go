package bot

import (
	"gopkg.in/telebot.v3"
	"log"
	"time"
)

type Bot struct {
	Users           []int64      // все пользователи здесь / их список
	b               *telebot.Bot // объект, который умеет общаться с тг апи
	alreadyNotified bool
}

// конструктор
// аналог b := &bot.Bot{}
func New() *Bot {
	return &Bot{}
}

// получаем chatID пользователя в tg куда будем отправлять
// запускаем в отдельной горутине, потому что
func (bot *Bot) Start(token string) {
	settings := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := telebot.NewBot(settings)
	if err != nil {
		log.Fatal(err)
	}
	b.Handle("/start", func(c telebot.Context) error {
		chatID := c.Sender().ID
		bot.Users = append(bot.Users, chatID)

		log.Printf("новый пользователь: %d", chatID)

		/*
			ИЗМЕНИТЬ RETURN
		*/

		return c.Send("Отлично, ты подписался на уведомления!")
	})
	bot.b = b // ПЕРЕДАЕМ ВСЕ В СТРУКТУРУ ЧТОБЫ ХРАНИЛ ЗНАЧЕНИЯ
	b.Start()
}

// отправляем уведомление в телеграм
func (bot *Bot) NotifyAll(message string) {
	if bot.alreadyNotified {
		return
	}

	for _, user := range bot.Users {
		rec := &telebot.User{ID: user} /*
			telebot не умеет отправлять просто по числу int64. Ему нужен объект который реализует интерфейс. Поэтому создаём минимальный telebot.User с одним полем ID — этого достаточно
		*/
		bot.b.Send(rec, message)
	}
}
