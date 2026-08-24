package alarm

import (
	"errors"
	"time"

	"pipewatch/internal/store"
)

// SectionSuppression is the flood-suppression state of one pipeline section.
type SectionSuppression struct {
	SectionID string    `json:"section_id"`
	Reason    string    `json:"reason"`
	Until     time.Time `json:"until"`
}

// Suppression tracks which sections hide repeat alarms during a flood.
type Suppression struct {
	entries map[string]*SectionSuppression
}

// NewSuppression creates an empty suppression set.
func NewSuppression() *Suppression {
	return &Suppression{entries: make(map[string]*SectionSuppression)}
}

// Start marks a section as flood-suppressed until the given time.
func (e *Engine) StartSuppression(sectionID, reason string, until time.Time) error {
	e.suppress.entries[sectionID] = &SectionSuppression{
		SectionID: sectionID,
		Reason:    reason,
		Until:     until,
	}
	return store.WriteJSON(e.st, "alarms/suppression/"+sectionID+".json", e.suppress.entries[sectionID])
}

// IsSuppressed reports whether a section is currently flood-suppressed.
func (e *Engine) IsSuppressed(sectionID string) bool {
	entry, ok := e.suppress.entries[sectionID]
	if !ok {
		return false
	}
	if time.Now().After(entry.Until) {
		delete(e.suppress.entries, sectionID)
		return false
	}
	return true
}

// EndSuppression clears the flood suppression of a section. The recovery
// record must already be durable before suppression ends.
func (e *Engine) EndSuppression(sectionID string, durable bool) error {
	if !durable {
		return errors.New("alarm: recovery record is not durable")
	}
	delete(e.suppress.entries, sectionID)
	return e.st.Remove("alarms/suppression/" + sectionID + ".json")
}

// ListSuppressions returns the active suppression entries.
func (e *Engine) ListSuppressions() []SectionSuppression {
	now := time.Now()
	var out []SectionSuppression
	for sectionID, entry := range e.suppress.entries {
		if now.After(entry.Until) {
			delete(e.suppress.entries, sectionID)
			continue
		}
		out = append(out, *entry)
	}
	return out
}
