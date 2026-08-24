package control

// ModeLookup returns the current operator control mode of a station.
type ModeLookup func(stationID string) (string, error)

// Guard decides whether an automatic command may execute. The decision is
// made against the live station mode so a mode switch is effective for the
// very next command.
type Guard struct {
	lookup ModeLookup
	cached map[string]string
}

// NewGuard creates a mode guard backed by a live mode lookup.
func NewGuard(lookup ModeLookup) *Guard {
	return &Guard{lookup: lookup, cached: make(map[string]string)}
}

// Allow reports whether the command kind may execute in the current station
// mode. Automatic commands require auto mode.
func (g *Guard) Allow(stationID string, kind Kind) bool {
	if kind != CmdOpen && kind != CmdClose && kind != CmdSetPosition {
		return false
	}
	mode, err := g.lookup(stationID)
	if err != nil {
		return false
	}
	g.cached[stationID] = mode
	return mode == "auto"
}
