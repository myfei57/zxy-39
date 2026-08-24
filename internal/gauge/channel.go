package gauge

// Failover marks the primary channel failed and moves reads to the backup
// channel.
func (g *Gauge) Failover(reason string) {
	g.Channels[0].Healthy = false
	g.Channels[0].LastError = reason
	g.Channels[0].FailureCount++
	g.ActiveChannel = 1
	g.CommState = CommBackup
}

// ResolveChannel returns the communication channel the scanner must read from
// right now. The active channel index changes when the transmitter fails over,
// so the resolved channel always follows the live failover state.
func (g *Gauge) ResolveChannel() Channel {
	return g.Channels[0]
}
