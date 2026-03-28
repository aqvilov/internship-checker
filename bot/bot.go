package bot

import (
	"internship/storage"
	"log"
	"time"

	telebot "gopkg.in/telebot.v3"
)

type Bot struct {
	b        *telebot.Bot
	storage  *storage.Storage
	notified map[string]bool // ключ типа bool чтобы чекать каждый сайт отдельно (отдельный флаг для каждого)
}

func New(storage *storage.Storage) *Bot {
	return &Bot{
		storage:  storage,
		notified: make(map[string]bool),
	}
}

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
		if err := bot.storage.AddUser(chatID); err != nil {
			log.Printf("ошибка сохранения пользователя: %v", err)
		} else {
			log.Printf("новый пользователь: %d", chatID)
		}
		return c.Send("Отлично, ты подписался на уведомления о стажировках!")
	})

	bot.b = b
	b.Start()
}

func (bot *Bot) NotifyAll(siteName string, message string) {
	if bot.notified[siteName] {
		return
	}

	users, err := bot.storage.GetUsers()
	if err != nil {
		log.Printf("ошибка получения пользователей: %v", err)
		return
	}

	if len(users) == 0 {
		log.Println("нет пользователей для рассылки")
		return
	}

	bot.notified[siteName] = true
	for _, chatID := range users {
		rec := &telebot.User{ID: chatID}
		if _, err := bot.b.Send(rec, message); err != nil {
			log.Printf("ошибка отправки пользователю %d: %v", chatID, err)
		}
	}
	log.Printf("разослано %d пользователям", len(users))
}
