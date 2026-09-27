package reports

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemoryStore struct {
	mu      sync.RWMutex
	reports []Report
	actions []Action
}

func (s *MemoryStore) List(_ context.Context, limit int) ([]Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]Report(nil), s.reports...)
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *MemoryStore) Transition(_ context.Context, id, actorID uuid.UUID, status string) (*Report, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.reports {
		if s.reports[index].ID == id {
			report := &s.reports[index]
			if !ValidTransition(report.Status, status) {
				return nil, ErrNotFound
			}
			actionID := uuid.New()
			s.actions = append(s.actions, Action{ID: actionID, ReportID: id, ActorID: actorID, FromStatus: report.Status, ToStatus: status, CreatedAt: time.Now().UTC()})
			report.Status = status
			report.UpdatedAt = time.Now().UTC()
			copy := *report
			return &copy, nil
		}
	}
	return nil, ErrNotFound
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{} }

func (s *MemoryStore) Create(_ context.Context, report Report) (*Report, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	report.CreatedAt = time.Now().UTC()
	report.UpdatedAt = report.CreatedAt
	s.reports = append(s.reports, report)
	copy := report
	return &copy, nil
}

func (s *MemoryStore) OpenCount(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, report := range s.reports {
		if report.Status != StatusResolved {
			count++
		}
	}
	return count, nil
}
