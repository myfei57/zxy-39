package historian

import (
	"time"

	"pipewatch/internal/gauge"
	"pipewatch/internal/store"
)

// Service accumulates readings into window summaries and persists them.
type Service struct {
	st      *store.Store
	windows *Windows
	pending map[string][]gauge.Reading
}

// NewService wires the historian to a store.
func NewService(st *store.Store) *Service {
	return &Service{
		st:      st,
		windows: NewWindows(),
		pending: make(map[string][]gauge.Reading),
	}
}

// AddReading queues a reading for the current window of its section.
func (s *Service) AddReading(r gauge.Reading) {
	if r.SectionID == "" {
		return
	}
	key := s.pendingKey(r.SectionID, WindowHour, r.TakenAt)
	s.pending[key] = append(s.pending[key], r)
}

func (s *Service) pendingKey(sectionID string, kind WindowKind, at time.Time) string {
	start := s.windows.Start(kind, at)
	return s.windows.Key(kind, start) + "|" + sectionID
}

// Aggregate computes the summary of the current window and persists it.
func (s *Service) Aggregate(sectionID string, kind WindowKind, at time.Time) (Summary, error) {
	start := s.windows.Start(kind, at)
	end := s.windows.End(start, kind)
	key := s.pendingKey(sectionID, kind, at)
	readings := s.pending[key]
	summary := Summary{
		ID:        sectionID + "-" + start.Format("2006010215"),
		SectionID: sectionID,
		Kind:      kind,
		Start:     start,
		End:       end,
	}
	count := 0
	for _, r := range readings {
		if r.TakenAt.Before(start) || !r.TakenAt.Before(end) {
			continue
		}
		count++
		summary.Sum += r.Value
		if count == 1 || r.Value < summary.Min {
			summary.Min = r.Value
		}
		if count == 1 || r.Value > summary.Max {
			summary.Max = r.Value
		}
	}
	summary.Count = count
	if summary.Count == 0 {
		return summary, nil
	}
	delete(s.pending, key)
	return summary, store.WriteJSON(s.st, s.windows.Key(kind, start), summary)
}
