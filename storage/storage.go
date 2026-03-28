package storage

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

// чтобы не делать поле db ПУБЛИЧНЫМ, МЫ СОЗДАЕМ ОТДЕЛЬНУЮ ФУНКЦИЮ
// ЕСЛИ ЧТО ЭТО ЮЗАЕМ В health/health.go
func (s *Storage) DB() *sql.DB {
	return s.db
}

func New(connStr string) (*Storage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            chat_id BIGINT PRIMARY KEY
        );
        CREATE TABLE IF NOT EXISTS subscriptions (
            chat_id   BIGINT,
            site_name TEXT,
            PRIMARY KEY (chat_id, site_name)
        );
    `)

	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) AddUser(chatID int64) error {
	_, err := s.db.Exec(`
		INSERT INTO users (chat_id) VALUES ($1)
		ON CONFLICT (chat_id) DO NOTHING
	`, chatID)
	return err
}

func (s *Storage) GetUsers() ([]int64, error) {
	rows, err := s.db.Query(`SELECT chat_id FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return nil, err
		}
		users = append(users, chatID)
	}
	return users, nil
}

// вместе мониторинга всех компаний, даем возможность подписаться на конкретные!
func (s *Storage) Subscribe(chatID int64, siteName string) error {
	_, err := s.db.Exec(`
        INSERT INTO subscriptions (chat_id, site_name) VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `, chatID, siteName)
	return err
}

func (s *Storage) Unsubscribe(chatID int64, siteName string) error {
	_, err := s.db.Exec(`
        DELETE FROM subscriptions WHERE chat_id = $1 AND site_name = $2
    `, chatID, siteName)
	return err
}

func (s *Storage) GetSubscriptions(chatID int64) ([]string, error) {
	rows, err := s.db.Query(`
        SELECT site_name FROM subscriptions WHERE chat_id = $1
    `, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		sites = append(sites, name)
	}
	return sites, nil
}

func (s *Storage) GetSubscribers(siteName string) ([]int64, error) {
	rows, err := s.db.Query(`
        SELECT chat_id FROM subscriptions WHERE site_name = $1
    `, siteName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var chatID int64
		rows.Scan(&chatID)
		users = append(users, chatID)
	}
	return users, nil
}
