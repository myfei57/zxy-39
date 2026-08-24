package gauge

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// ErrTimeout is returned when a transmitter does not answer within the scan
// budget.
var ErrTimeout = errors.New("gauge read timeout")

// Source adapts a transmitter driver to a gauge. The console and tests may
// install their own source; production wiring uses the demo driver.
type Source interface {
	Read(g *Gauge) (float64, error)
}

// SourceFunc adapts a plain function to the Source interface.
type SourceFunc func(g *Gauge) (float64, error)

// Read implements Source.
func (f SourceFunc) Read(g *Gauge) (float64, error) {
	return f(g)
}

type demoSource struct{}

// DemoSource returns the production transmitter driver.
func DemoSource() Source {
	return demoSource{}
}

// Read returns a deterministic value that depends on the resolved channel.
func (demoSource) Read(g *Gauge) (float64, error) {
	base := 6.0
	if g.ResolveChannel().Name == "B" {
		base = 8.0
	}
	phase := float64(time.Now().UnixNano()/int64(time.Second)%60) * math.Pi / 30
	return base + math.Sin(phase)*0.05, nil
}

// Read samples one transmitter through its current channel and records the
// quality transition on the gauge.
func (r *Registry) Read(ctx context.Context, gaugeID string) (Reading, error) {
	g, ok := r.Get(gaugeID)
	if !ok {
		return Reading{}, fmt.Errorf("gauge: %s not found", gaugeID)
	}
	ch := g.ResolveChannel()
	if !ch.Healthy {
		g.LastQuality = QualityStale
		_ = r.Save(g)
		return Reading{}, fmt.Errorf("gauge: channel %s unhealthy: %w", ch.Name, ErrTimeout)
	}
	value, err := r.source.Read(g)
	if err != nil {
		g.LastQuality = QualityFailed
		_ = r.Save(g)
		return Reading{}, err
	}
	now := time.Now()
	reading := Reading{
		GaugeID:   g.ID,
		StationID: g.StationID,
		Kind:      g.Kind,
		Value:     value,
		Raw:       value,
		Quality:   QualityGood,
		TakenAt:   now,
	}
	g.RawValue = value
	g.LastQuality = QualityGood
	g.LastReading = now
	_ = r.Save(g)
	return reading, nil
}
