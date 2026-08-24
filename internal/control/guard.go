package control

// ModeLookup returns the current operator control mode of a station.
type ModeLookup func(stationID string) (string, error)

// Guard decides whether an automatic command may execute. The decision is
// made against the live station mode so a mode switch is effective for the
// very next command.
type Guard struct {
	lookup ModeLookup
}

// NewGuard creates a mode guard backed by a live mode lookup.
func NewGuard(lookup ModeLookup) *Guard {
	return &Guard{lookup: lookup}
}

// Allow reports whether the command kind may execute in the current station
// mode. Automatic commands require auto mode. The live mode is read on every
// call so a switch to manual takes effect immediately and never serves a stale
// cached mode.
func (g *Guard) Allow(stationID string, kind Kind) bool {
	if kind != CmdOpen && kind != CmdClose && kind != CmdSetPosition {
		return false
	}
	mode, err := g.lookup(stationID)
	if err != nil {
		return false
	}
	return mode == "auto"
}
