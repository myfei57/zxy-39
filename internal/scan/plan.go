package scan

import "pipewatch/internal/gauge"

// SetPlan fixes the scan plan and the sections that own the gauges.
func (s *Service) SetPlan(plan []*gauge.Gauge, sections []string) {
	s.plan = plan
	s.sections = sections
}
