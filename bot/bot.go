package bot

import (
	"internship/checker"
	"internship/storage"
	"log"
	"strings"
	"time"

	telebot "gopkg.in/telebot.v3"
)

type Bot struct {
	b        *telebot.Bot
	storage  *storage.Storage
	notified map[string]bool
	sites    []checker.Site // список всех сайтов для кнопок
}

func New(storage *storage.Storage, sites []checker.Site) *Bot {
	return &Bot{
		storage:  storage,
		notified: make(map[string]bool),
		sites:    sites,
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
		}
		return bot.sendMenu(c, chatID)
	})

	// обработчик нажатия на кнопку
	b.Handle(telebot.OnCallback, func(c telebot.Context) error {
		chatID := c.Sender().ID
		data := c.Callback().Data

		// получаем текущие подписки на стажки
		subs, err := bot.storage.GetSubscriptions(chatID)
		if err != nil {
			return c.Respond(&telebot.CallbackResponse{Text: "ошибка!"})
		}

		// проверяем
		isSubscribed := false
		for _, s := range subs {
			if s == data {
				isSubscribed = true
				break
			}
		}

		if isSubscribed {
			bot.storage.Unsubscribe(chatID, data)
		} else {
			bot.storage.Subscribe(chatID, data)
		}

		// обновляем меню
		if err := bot.editMenu(c, chatID); err != nil {
			log.Printf("ошибка обновления меню: %v", err)
		}
		return c.Respond(&telebot.CallbackResponse{})
	})

	bot.b = b
	b.Start()
}

// кнопочки
func (bot *Bot) sendMenu(c telebot.Context, chatID int64) error {
	markup, text, err := bot.buildMenu(chatID)
	if err != nil {
		return err
	}
	return c.Send(text, markup)
}

func (bot *Bot) editMenu(c telebot.Context, chatID int64) error {
	markup, text, err := bot.buildMenu(chatID)
	if err != nil {
		return err
	}
	return c.Edit(text, markup)
}

// текст и кнопки с учётом текущих подписок
func (bot *Bot) buildMenu(chatID int64) (*telebot.ReplyMarkup, string, error) {
	subs, err := bot.storage.GetSubscriptions(chatID)
	if err != nil {
		return nil, "", err
	}

	subMap := make(map[string]bool)
	for _, s := range subs {
		subMap[s] = true
	}

	markup := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	for _, site := range bot.sites {
		label := site.Name
		if subMap[site.Name] {
			label = " " + site.Name
		}
		btn := markup.Data(label, site.Name, site.Name)
		rows = append(rows, markup.Row(btn))
	}
	markup.Inline(rows...)

	text := "Выбери компании, у которых будем отслеживать стажировки:\nВторое нажатие по одной кнопке - прекращения отслеживание этой компании\n\n"
	if len(subs) == 0 {
		text += "В данный момент выбрано: Сейчас ничего не выбрано"
	} else {
		text += strings.Join(subs, ", ")
	}

	return markup, text, nil
}

func (bot *Bot) NotifyAll(siteName string, message string) {
	if bot.notified[siteName] {
		return
	}

	users, err := bot.storage.GetSubscribers(siteName)
	if err != nil {
		log.Printf("ошибка получения подписчиков: %v", err)
		return
	}

	if len(users) == 0 {
		log.Printf("нет подписчиков на %s", siteName)
		return
	}

	bot.notified[siteName] = true
	for _, chatID := range users {
		rec := &telebot.User{ID: chatID}
		if _, err := bot.b.Send(rec, message); err != nil {
			log.Printf("ошибка отправки пользователю %d: %v", chatID, err)
		}
	}
	log.Printf("%s: разослано %d пользователям", siteName, len(users))
}
