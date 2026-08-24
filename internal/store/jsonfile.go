package store

import (
	"encoding/json"
	"fmt"
)

// WriteJSON marshals v and persists it at rel.
func WriteJSON(s *Store, rel string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("store: marshal %s: %w", rel, err)
	}
	return s.WriteAtomic(rel, data)
}

// ReadJSON loads rel and unmarshals it into v.
func ReadJSON(s *Store, rel string, v any) error {
	data, err := s.Read(rel)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("store: unmarshal %s: %w", rel, err)
	}
	return nil
}
