package user

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
)

func TestMemoryRepositoryCreate(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	user := User{
		Name: "Alice",
		Age:  30,
	}

	rUser, err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("user created with error %v", err)
	}

	wantID := 1
	if rUser.ID != wantID {
		t.Errorf("user created with id = %d; want id = %d", rUser.ID, wantID)
	}
	if rUser.Name != user.Name {
		t.Errorf("create user with name = % q; want name = %q", rUser.Name, user.Name)
	}
	if rUser.Age != user.Age {
		t.Errorf("create user with age = %d; want age = %d", rUser.Age, user.Age)
	}

	secondUser, err := repo.Create(ctx, User{
		Name: "Bob",
		Age:  42,
	})
	if err != nil {
		t.Fatalf("user created with error %v", err)
	}
	nextID := 2
	if secondUser.ID != nextID {
		t.Errorf("nextID is = %d; want = %d", repo.nextID, nextID)
	}
}

func TestMemoryRepositoryByID(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	user := User{
		Name: "Bob",
		Age:  44,
	}

	cUser, err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("user created with error %v", err)
	}

	userID := cUser.ID
	rUser, err := repo.ByID(ctx, userID)

	if err != nil {
		t.Fatalf("got user with error %v", err)
	}

	if rUser.Name != user.Name {
		t.Errorf("got user with name = %q; want name = %q", rUser.Name, user.Name)
	}
	if rUser.Age != user.Age {
		t.Errorf("got user with age = %d; want age = %d", rUser.Age, user.Age)
	}
}

func TestMemoryRepositoryNotFound(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	userID := 999
	_, err := repo.ByID(ctx, userID)

	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got user with error %v; want error = %v", err, ErrNotFound)
	}
}

func TestMemoryRepositoryConcurrentCreate(t *testing.T) {
	count := 1000
	ctx := context.Background()
	repo := NewMemoryRepository()
	var wg sync.WaitGroup
	for i := 1; i <= count; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := repo.Create(ctx, User{
				Name: "Name_" + strconv.Itoa(i),
				Age:  i,
			})

			if err != nil {
				t.Errorf("got error %v", err)
			}
		}()
	}

	wg.Wait()

	for i := 1; i <= count; i++ {
		_, err := repo.ByID(ctx, i)
		if errors.Is(err, ErrNotFound) {
			t.Fatalf("user %d not found", i)
		}
		if err != nil {
			t.Fatalf("got error %v", err)
		}
	}
}
