package user

import (
	"context"
	"sync"
)

type MemoryRepository struct {
	mu     sync.RWMutex
	users  map[int]User
	nextID int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:  make(map[int]User),
		nextID: 1,
	}
}

func (r *MemoryRepository) Create(ctx context.Context, user User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.ID = r.nextID

	r.users[r.nextID] = user
	r.nextID++

	return user, nil
}

func (r *MemoryRepository) ByID(ctx context.Context, id int) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}
