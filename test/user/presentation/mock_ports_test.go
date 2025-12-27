package presentation_test

import "GAMERS-BE/internal/user/domain"

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
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserQueryPort) FindByEmail(email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
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
		return domain.ErrUserNotFound
	}
	m.queryPort.users[user.Id] = user
	return nil
}

func (m *mockUserCommandPort) DeleteById(id int64) error {
	if _, exists := m.queryPort.users[id]; !exists {
		return domain.ErrUserNotFound
	}
	delete(m.queryPort.users, id)
	return nil
}

type mockProfileQueryPort struct {
	profiles map[int64]*domain.Profile
}

func newMockProfileQueryPort() *mockProfileQueryPort {
	return &mockProfileQueryPort{
		profiles: make(map[int64]*domain.Profile),
	}
}

func (m *mockProfileQueryPort) FindById(id int64) (*domain.Profile, error) {
	profile, exists := m.profiles[id]
	if !exists {
		return nil, domain.ErrProfileNotFound
	}
	return profile, nil
}

func (m *mockProfileQueryPort) FindByUserId(userId int64) (*domain.Profile, error) {
	for _, profile := range m.profiles {
		if profile.UserId == userId {
			return profile, nil
		}
	}
	return nil, domain.ErrProfileNotFound
}

type mockProfileCommandPort struct {
	queryPort *mockProfileQueryPort
	nextID    int64
}

func newMockProfileCommandPort(queryPort *mockProfileQueryPort) *mockProfileCommandPort {
	return &mockProfileCommandPort{
		queryPort: queryPort,
		nextID:    1,
	}
}

func (m *mockProfileCommandPort) Save(profile *domain.Profile) error {
	if profile.Id == 0 {
		profile.Id = m.nextID
		m.nextID++
	}
	m.queryPort.profiles[profile.Id] = profile
	return nil
}

func (m *mockProfileCommandPort) Update(profile *domain.Profile) error {
	if _, exists := m.queryPort.profiles[profile.Id]; !exists {
		return domain.ErrProfileNotFound
	}
	m.queryPort.profiles[profile.Id] = profile
	return nil
}

func (m *mockProfileCommandPort) DeleteById(id int64) error {
	if _, exists := m.queryPort.profiles[id]; !exists {
		return domain.ErrProfileNotFound
	}
	delete(m.queryPort.profiles, id)
	return nil
}
