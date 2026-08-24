// Package store provides the file-backed persistence layer shared by all
// PipeWatch components. Documents are stored as JSON files under a root data
// directory; every write goes through a temp-file-and-rename cycle so a
// partial write never corrupts an existing document.
package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Store is a rooted file-backed document store.
type Store struct {
	root string
}

// New creates a Store rooted at root, creating the directory when needed.
func New(root string) (*Store, error) {
	if root == "" {
		return nil, fmt.Errorf("store: empty root")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("store: create root %q: %w", root, err)
	}
	return &Store{root: root}, nil
}

// Join resolves a relative document path below the store root.
func (s *Store) Join(parts ...string) string {
	return filepath.Join(append([]string{s.root}, parts...)...)
}

// WriteAtomic persists data at rel, replacing any previous document.
func (s *Store) WriteAtomic(rel string, data []byte) error {
	target := s.Join(rel)
	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("store: create parent for %s: %w", rel, err)
	}
	if err := writeFileSync(target, data); err != nil {
		return fmt.Errorf("store: write %s: %w", rel, err)
	}
	return nil
}

// Read loads the document at rel.
func (s *Store) Read(rel string) ([]byte, error) {
	data, err := os.ReadFile(s.Join(rel))
	if err != nil {
		return nil, fmt.Errorf("store: read %s: %w", rel, err)
	}
	return data, nil
}

// Remove deletes the document at rel.
func (s *Store) Remove(rel string) error {
	if err := os.Remove(s.Join(rel)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("store: remove %s: %w", rel, err)
	}
	return nil
}

// List returns the relative file paths below dir, sorted.
func (s *Store) List(dir string) ([]string, error) {
	full := s.Join(dir)
	entries, err := os.ReadDir(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("store: list %s: %w", dir, err)
	}
	var out []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		out = append(out, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(out)
	return out, nil
}

// AppendLine appends a single line to the document at rel, creating it when
// needed.
func (s *Store) AppendLine(rel string, line []byte) error {
	target := s.Join(rel)
	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("store: create parent for %s: %w", rel, err)
	}
	f, err := os.OpenFile(target, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("store: append %s: %w", rel, err)
	}
	defer f.Close()
	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("store: append %s: %w", rel, err)
	}
	return nil
}
