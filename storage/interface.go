package storage

// интерфейс для мок тестирования /storage
type IStorage interface {
	AddUser(chatID int64) error
	GetUsers() ([]int64, error)
	Subscribe(chatID int64, siteName string) error
	Unsubscribe(chatID int64, siteName string) error
	GetSubscriptions(chatID int64) ([]string, error)
	GetSubscribers(siteName string) ([]int64, error)
}
