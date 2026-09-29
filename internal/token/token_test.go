package token

import (
	"errors"
	"testing"
	"time"
)

func TestManager_GenerateAndVerify(t *testing.T) {
	var tests = []struct {
		name        string
		secret      string
		ttl         time.Duration
		modifSecret string
		wait        time.Duration
		isCorrupt   bool
		userID      int
		wantErr     error
	}{
		{
			name:        "modified secret",
			secret:      "12345",
			ttl:         10 * time.Minute,
			modifSecret: "54321",
			userID:      123,
			wantErr:     ErrInvalidToken,
		},
		{
			name:    "modified ttl",
			secret:  "12345",
			ttl:     1 * time.Second,
			wait:    2 * time.Second,
			userID:  123,
			wantErr: ErrInvalidToken,
		},
		{
			name:      "modified token",
			secret:    "12345",
			ttl:       10 * time.Minute,
			userID:    123,
			isCorrupt: true,
			wantErr:   ErrInvalidToken,
		},
		{
			name:   "success",
			secret: "12345",
			ttl:    10 * time.Minute,
			userID: 123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager([]byte(tt.secret), tt.ttl)
			token, err := manager.Generate(tt.userID)
			if err != nil {
				t.Fatalf("generate token %v", err)
			}
			if token == "" {
				t.Fatal("token can not be empty")
			}

			if tt.modifSecret != "" {
				manager.secret = []byte(tt.modifSecret)
			}
			if tt.isCorrupt {
				token = token + "+"
			}
			if tt.wait > 0 {
				time.Sleep(tt.wait)
			}
			userID, err := manager.Verify(token)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v; got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("verify token %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected Verify() error: %v", err)
			}

			if userID != tt.userID {
				t.Errorf("got userID %d; want %d", userID, tt.userID)
			}
		})
	}
}
