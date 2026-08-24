package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
)

// List returns the events of one kind recorded at or after from.
func (s *Service) List(kind Kind, from time.Time) ([]Event, error) {
	all, err := s.readAll()
	if err != nil {
		return nil, err
	}
	var out []Event
	for _, e := range all {
		if e.Kind == kind && !e.At.Before(from) {
			out = append(out, e)
		}
	}
	return out, nil
}

// Recent returns the newest n events across all kinds.
func (s *Service) Recent(n int) ([]Event, error) {
	all, err := s.readAll()
	if err != nil {
		return nil, err
	}
	if len(all) > n {
		all = all[len(all)-n:]
	}
	return all, nil
}

func (s *Service) readAll() ([]Event, error) {
	f, err := os.Open(s.st.Join("audit/events.jsonl"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var out []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, scanner.Err()
}
