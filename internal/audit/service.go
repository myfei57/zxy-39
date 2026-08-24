package audit

import "pipewatch/internal/store"

// Service owns the audit journal.
type Service struct {
	st *store.Store
}

// NewService creates an audit service over a store.
func NewService(st *store.Store) *Service {
	return &Service{st: st}
}

// Load prepares the journal directory.
func (s *Service) Load() error {
	return s.st.WriteAtomic("audit/.keep", []byte(""))
}
