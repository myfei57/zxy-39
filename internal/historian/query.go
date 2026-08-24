package historian

import (
	"time"

	"pipewatch/internal/store"
)

// Query returns the persisted summaries of a section inside [from, to).
func (s *Service) Query(sectionID string, kind WindowKind, from, to time.Time) ([]Summary, error) {
	files, err := s.st.List("historian/summaries/" + string(kind))
	if err != nil {
		return nil, err
	}
	var out []Summary
	for _, rel := range files {
		var summary Summary
		if err := store.ReadJSON(s.st, rel, &summary); err != nil {
			return nil, err
		}
		if summary.SectionID != sectionID {
			continue
		}
		if summary.Start.Before(from) || !summary.Start.Before(to) {
			continue
		}
		out = append(out, summary)
	}
	return out, nil
}

// List returns every persisted summary of a section.
func (s *Service) List(sectionID string, kind WindowKind) ([]Summary, error) {
	files, err := s.st.List("historian/summaries/" + string(kind))
	if err != nil {
		return nil, err
	}
	var out []Summary
	for _, rel := range files {
		var summary Summary
		if err := store.ReadJSON(s.st, rel, &summary); err != nil {
			return nil, err
		}
		if summary.SectionID == sectionID {
			out = append(out, summary)
		}
	}
	return out, nil
}
