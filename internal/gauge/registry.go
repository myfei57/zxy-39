package gauge

import (
	"fmt"

	"pipewatch/internal/store"
)

// Registry owns the transmitter directory, persisted under gauges/.
type Registry struct {
	st     *store.Store
	gauges map[string]*Gauge
	order  []string
	source Source
}

// NewRegistry creates an empty transmitter registry with the demo driver.
func NewRegistry(st *store.Store) *Registry {
	return &Registry{
		st:     st,
		gauges: make(map[string]*Gauge),
		source: demoSource{},
	}
}

// SetSource installs the transmitter driver used by Read.
func (r *Registry) SetSource(src Source) {
	r.source = src
}

// Register stores a new transmitter document.
func (r *Registry) Register(g *Gauge) error {
	if g.ID == "" || g.StationID == "" || g.Tag == "" {
		return fmt.Errorf("gauge: id, station and tag are required")
	}
	if _, exists := r.gauges[g.ID]; !exists {
		r.order = append(r.order, g.ID)
	}
	r.gauges[g.ID] = g
	return r.Save(g)
}

// Get returns the transmitter with the given id.
func (r *Registry) Get(id string) (*Gauge, bool) {
	g, ok := r.gauges[id]
	return g, ok
}

// List returns all transmitters in registration order.
func (r *Registry) List() []*Gauge {
	var out []*Gauge
	for _, id := range r.order {
		if g, ok := r.gauges[id]; ok {
			out = append(out, g)
		}
	}
	return out
}

// Save persists one transmitter document.
func (r *Registry) Save(g *Gauge) error {
	return store.WriteJSON(r.st, "gauges/"+g.ID+".json", g)
}

// Failover switches a transmitter to its backup channel and persists the new
// communication state.
func (r *Registry) Failover(gaugeID, reason string) error {
	g, ok := r.gauges[gaugeID]
	if !ok {
		return fmt.Errorf("gauge: %s not found", gaugeID)
	}
	g.Failover(reason)
	return r.Save(g)
}

// Load restores transmitters from the store.
func (r *Registry) Load() error {
	files, err := r.st.List("gauges")
	if err != nil {
		return err
	}
	for _, rel := range files {
		var g Gauge
		if err := store.ReadJSON(r.st, rel, &g); err != nil {
			return err
		}
		r.Register(&g)
	}
	return nil
}
