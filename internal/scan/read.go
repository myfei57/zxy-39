package scan

import (
	"context"
	"fmt"

	"pipewatch/internal/gauge"
)

// Poller reads gauges through the transmitter registry. The active channel is
// resolved on every read so a communication failover takes effect
// immediately.
type Poller struct {
	gauges *gauge.Registry
	active map[string]int
}

// NewPoller creates a scanner reader over a transmitter registry.
func NewPoller(gauges *gauge.Registry) *Poller {
	p := &Poller{gauges: gauges, active: make(map[string]int)}
	for _, g := range gauges.List() {
		p.active[g.ID] = g.ActiveChannel
	}
	return p
}

// ReadOne samples one gauge through its current channel.
func (p *Poller) ReadOne(ctx context.Context, gaugeID string) (gauge.Reading, error) {
	g, ok := p.gauges.Get(gaugeID)
	if !ok {
		return gauge.Reading{}, fmt.Errorf("scan: unknown gauge %s", gaugeID)
	}
	if index, ok := p.active[g.ID]; ok {
		g.ActiveChannel = index
	}
	return p.gauges.Read(ctx, g.ID)
}
