package storage

//тупо реализуем интерфейс

type MockStorage struct {
	users         []int64
	subscriptions map[int64][]string
}

func NewMock() *MockStorage {
	return &MockStorage{
		subscriptions: make(map[int64][]string),
	}
}

func (m *MockStorage) AddUser(chatID int64) error {
	m.users = append(m.users, chatID)
	return nil
}

func (m *MockStorage) GetUsers() ([]int64, error) {
	return m.users, nil
}

func (m *MockStorage) Subscribe(chatID int64, siteName string) error {
	m.subscriptions[chatID] = append(m.subscriptions[chatID], siteName)
	return nil
}

func (m *MockStorage) Unsubscribe(chatID int64, siteName string) error {
	subs := m.subscriptions[chatID]
	for i, s := range subs {
		if s == siteName {
			m.subscriptions[chatID] = append(subs[:i], subs[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockStorage) GetSubscriptions(chatID int64) ([]string, error) {
	return m.subscriptions[chatID], nil
}

func (m *MockStorage) GetSubscribers(siteName string) ([]int64, error) {
	var result []int64
	for chatID, subs := range m.subscriptions {
		for _, s := range subs {
			if s == siteName {
				result = append(result, chatID)
			}
		}
	}
	return result, nil
}
