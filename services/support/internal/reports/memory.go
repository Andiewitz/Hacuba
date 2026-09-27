package reports

import (
	"context"
	"sync"
	"time"
)

type MemoryStore struct {
	mu      sync.RWMutex
	reports []Report
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{} }

func (s *MemoryStore) Create(_ context.Context, report Report) (*Report, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	report.CreatedAt = time.Now().UTC()
	s.reports = append(s.reports, report)
	copy := report
	return &copy, nil
}

func (s *MemoryStore) OpenCount(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, report := range s.reports {
		if report.Status == StatusOpen {
			count++
		}
	}
	return count, nil
}
