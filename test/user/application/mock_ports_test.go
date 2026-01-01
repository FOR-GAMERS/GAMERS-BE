package application_test

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/user/domain"
)

type mockUserQueryPort struct {
	users map[int64]*domain.User
}

func newMockUserQueryPort() *mockUserQueryPort {
	return &mockUserQueryPort{
		users: make(map[int64]*domain.User),
	}
}

func (m *mockUserQueryPort) FindById(id int64) (*domain.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, exception.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserQueryPort) FindByEmail(email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, exception.ErrUserNotFound
}

type mockUserCommandPort struct {
	queryPort *mockUserQueryPort
	nextID    int64
}

func newMockUserCommandPort(queryPort *mockUserQueryPort) *mockUserCommandPort {
	return &mockUserCommandPort{
		queryPort: queryPort,
		nextID:    1,
	}
}

func (m *mockUserCommandPort) Save(user *domain.User) error {
	if user.Id == 0 {
		user.Id = m.nextID
		m.nextID++
	}
	m.queryPort.users[user.Id] = user
	return nil
}

func (m *mockUserCommandPort) Update(user *domain.User) error {
	if _, exists := m.queryPort.users[user.Id]; !exists {
		return exception.ErrUserNotFound
	}
	m.queryPort.users[user.Id] = user
	return nil
}

func (m *mockUserCommandPort) DeleteById(id int64) error {
	if _, exists := m.queryPort.users[id]; !exists {
		return exception.ErrUserNotFound
	}
	delete(m.queryPort.users, id)
	return nil
}
