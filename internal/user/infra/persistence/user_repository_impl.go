package persistence

import (
	"GAMERS-BE/internal/user/domain"
	"sync"
)

type InMemoryUserRepository struct {
	mu       sync.RWMutex
	users    map[int64]*domain.User
	nextID   int64
	emailIdx map[string]int64
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:    make(map[int64]*domain.User),
		nextID:   1,
		emailIdx: make(map[string]int64),
	}
}

func (r *InMemoryUserRepository) Save(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 이메일 중복 체크
	if existingID, exists := r.emailIdx[user.Email]; exists {
		if existingID != user.Id {
			return domain.ErrUserAlreadyExists
		}
	}

	// 새 사용자인 경우 ID 할당
	if user.Id == 0 {
		user.Id = r.nextID
		r.nextID++
	}

	r.users[user.Id] = user
	r.emailIdx[user.Email] = user.Id

	return nil
}

func (r *InMemoryUserRepository) FindById(id int64) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	userCopy := *user
	return &userCopy, nil
}

func (r *InMemoryUserRepository) Update(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Id]; !exists {
		return domain.ErrUserNotFound
	}

	if existingID, exists := r.emailIdx[user.Email]; exists {
		if existingID != user.Id {
			return domain.ErrUserAlreadyExists
		}
	}

	if oldUser, exists := r.users[user.Id]; exists {
		delete(r.emailIdx, oldUser.Email)
	}

	r.users[user.Id] = user
	r.emailIdx[user.Email] = user.Id

	return nil
}

func (r *InMemoryUserRepository) DeleteById(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return domain.ErrUserNotFound
	}

	delete(r.emailIdx, user.Email)
	delete(r.users, id)

	return nil
}
