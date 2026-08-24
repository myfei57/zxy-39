package audit

import (
	"encoding/json"
)

// Record appends one audit event to the journal.
func (s *Service) Record(e Event) error {
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return s.st.AppendLine("audit/events.jsonl", append(line, '\n'))
}
