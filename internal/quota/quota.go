package quota

import (
	"errors"
	"fmt"

	"pipewatch/internal/store"
)

// ErrQuotaExceeded is returned when a section already stored its full reading
// budget.
var ErrQuotaExceeded = errors.New("quota: section reading budget exceeded")

// Quota is the reading budget of one pipeline section.
type Quota struct {
	SectionID   string `json:"section_id"`
	MaxReadings int    `json:"max_readings"`
	Used        int    `json:"used"`
}

// Service checks and consumes per-section reading budgets.
type Service struct {
	st     *store.Store
	quotas map[string]*Quota
}

// NewService creates an empty quota service.
func NewService(st *store.Store) *Service {
	return &Service{
		st:     st,
		quotas: make(map[string]*Quota),
	}
}

// Check verifies that a section still has reading budget left.
func (s *Service) Check(sectionID string) error {
	q, ok := s.quotas[sectionID]
	if !ok {
		return fmt.Errorf("quota: no budget configured for %s", sectionID)
	}
	if q.Used >= q.MaxReadings {
		return ErrQuotaExceeded
	}
	return nil
}

// Consume books one reading against the section budget.
func (s *Service) Consume(sectionID string) error {
	q, ok := s.quotas[sectionID]
	if !ok {
		return fmt.Errorf("quota: no budget configured for %s", sectionID)
	}
	if q.Used >= q.MaxReadings {
		return ErrQuotaExceeded
	}
	q.Used++
	return store.WriteJSON(s.st, "quota/"+sectionID+".json", q)
}

// Set configures or replaces the reading budget of a section.
func (s *Service) Set(sectionID string, maxReadings int) error {
	if maxReadings < 0 {
		return fmt.Errorf("quota: negative budget")
	}
	q, ok := s.quotas[sectionID]
	if !ok {
		q = &Quota{SectionID: sectionID}
		s.quotas[sectionID] = q
	}
	q.MaxReadings = maxReadings
	if q.Used > maxReadings {
		q.Used = maxReadings
	}
	return store.WriteJSON(s.st, "quota/"+sectionID+".json", q)
}

// Get returns the budget of a section.
func (s *Service) Get(sectionID string) (Quota, bool) {
	q, ok := s.quotas[sectionID]
	if !ok {
		return Quota{}, false
	}
	return *q, true
}

// List returns all configured budgets.
func (s *Service) List() []Quota {
	out := make([]Quota, 0, len(s.quotas))
	for _, q := range s.quotas {
		out = append(out, *q)
	}
	return out
}

// Load restores quotas from the store.
func (s *Service) Load() error {
	files, err := s.st.List("quota")
	if err != nil {
		return err
	}
	for _, rel := range files {
		var q Quota
		if err := store.ReadJSON(s.st, rel, &q); err != nil {
			return err
		}
		s.quotas[q.SectionID] = &q
	}
	return nil
}
