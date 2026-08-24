package ns

import (
	"fmt"

	"pipewatch/internal/store"
)

// Registry owns the pipeline and segment namespace, persisted under ns/.
type Registry struct {
	st        *store.Store
	pipelines map[string]*Pipeline
	segments  map[string]*Segment
	order     []string
}

// NewRegistry creates an empty namespace registry.
func NewRegistry(st *store.Store) *Registry {
	return &Registry{
		st:        st,
		pipelines: make(map[string]*Pipeline),
		segments:  make(map[string]*Segment),
	}
}

// RegisterPipeline stores a new pipeline document.
func (r *Registry) RegisterPipeline(p *Pipeline) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if _, exists := r.pipelines[p.ID]; !exists {
		r.order = append(r.order, p.ID)
	}
	r.pipelines[p.ID] = p
	return store.WriteJSON(r.st, "ns/pipelines/"+p.ID+".json", p)
}

// RegisterSegment stores a new segment document.
func (r *Registry) RegisterSegment(s *Segment) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if _, exists := r.pipelines[s.PipelineID]; !exists {
		return fmt.Errorf("ns: pipeline %s not registered", s.PipelineID)
	}
	r.segments[s.ID] = s
	return store.WriteJSON(r.st, "ns/segments/"+s.ID+".json", s)
}

// ListPipelines returns all pipelines in registration order.
func (r *Registry) ListPipelines() []*Pipeline {
	var out []*Pipeline
	for _, id := range r.order {
		if p, ok := r.pipelines[id]; ok {
			out = append(out, p)
		}
	}
	return out
}

// Load restores pipelines and segments from the store.
func (r *Registry) Load() error {
	files, err := r.st.List("ns/pipelines")
	if err != nil {
		return err
	}
	for _, rel := range files {
		var p Pipeline
		if err := store.ReadJSON(r.st, rel, &p); err != nil {
			return err
		}
		r.RegisterPipeline(&p)
	}
	segFiles, err := r.st.List("ns/segments")
	if err != nil {
		return err
	}
	for _, rel := range segFiles {
		var s Segment
		if err := store.ReadJSON(r.st, rel, &s); err != nil {
			return err
		}
		r.segments[s.ID] = &s
	}
	return nil
}
