package repository

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/bcrypt"

	"redditclone/internal/domain"
)

type UserRepository interface {
	Create(user *domain.User) error
	GetByUsername(username string) (*domain.User, error)
	GetByID(id string) (*domain.User, error)
}

type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
	seq   uint64
}

func NewMemoryUserRepository() *MemoryUserRepository {
	repo := &MemoryUserRepository{
		users: make(map[string]*domain.User),
	}
	repo.seedData()
	return repo
}

func (r *MemoryUserRepository) Create(u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Case-insensitive check for username uniqueness
	for _, existing := range r.users {
		if strings.EqualFold(existing.Username, u.Username) {
			return domain.ErrUserAlreadyExists
		}
	}

	idNum := atomic.AddUint64(&r.seq, 1) + 100
	u.ID = fmt.Sprintf("user_%d", idNum)

	r.users[u.ID] = u
	return nil
}

func (r *MemoryUserRepository) GetByUsername(username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if strings.EqualFold(u.Username, username) {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *MemoryUserRepository) GetByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *MemoryUserRepository) seedData() {
	hash1, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	hash2, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user1 := &domain.User{
		ID:       "507f1f77bcf86cd799439011",
		Username: "alice",
		Password: string(hash1),
	}
	user2 := &domain.User{
		ID:       "507f1f77bcf86cd799439012",
		Username: "bob",
		Password: string(hash2),
	}

	r.users[user1.ID] = user1
	r.users[user2.ID] = user2
}
