package services

import (
	"sync"
	"time"
)

type SmsStorage interface {
	Get(phoneNumber string) (string, bool)
	Set(phoneNumber string, code string)
}

type ThreadSafeSmsStorage struct {
	mutex sync.RWMutex

	// format: phone_number: {code: "9876", expiresAt: time.Time}
	storage map[string]smsEntry
}

type smsEntry struct {
	code      string
	expiresAt time.Time
}

func NewSmsStorage() *ThreadSafeSmsStorage {
	return &ThreadSafeSmsStorage{
		storage: make(map[string]smsEntry),
	}
}

func (s *ThreadSafeSmsStorage) Set(phoneNumber, code string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.storage[phoneNumber] = smsEntry{
		code:      code,
		expiresAt: time.Now().Add(3 * time.Minute),
	}
}

func (s *ThreadSafeSmsStorage) Get(phoneNumber string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entry, exists := s.storage[phoneNumber]
	if !exists {
		return "", false
	}

	if time.Now().After(entry.expiresAt) {
		s.cleanup()
		return "", false
	}

	return entry.code, true
}

// cleaning expired notes (not so fast, but sometimes)
func (s *ThreadSafeSmsStorage) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for phone, entry := range s.storage {
			if now.After(entry.expiresAt) {
				delete(s.storage, phone)
			}
		}
		s.mutex.Unlock()
	}
}
