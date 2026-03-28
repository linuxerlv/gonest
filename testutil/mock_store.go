package testutil

import (
	"context"
	"time"
)

type MockTokenStore struct {
	tokens map[string]*TokenData
}

type TokenData struct {
	UserID    string
	Username  string
	Roles     []string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewMockTokenStore() *MockTokenStore {
	return &MockTokenStore{
		tokens: make(map[string]*TokenData),
	}
}

func (s *MockTokenStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	s.tokens[key] = &TokenData{
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (s *MockTokenStore) Get(ctx context.Context, key string) (any, error) {
	data, exists := s.tokens[key]
	if !exists {
		return nil, nil
	}
	return data, nil
}

func (s *MockTokenStore) Delete(ctx context.Context, key string) error {
	delete(s.tokens, key)
	return nil
}

func (s *MockTokenStore) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := s.tokens[key]
	return exists, nil
}

func (s *MockTokenStore) Clear() {
	s.tokens = make(map[string]*TokenData)
}
