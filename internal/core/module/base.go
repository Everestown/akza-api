package module

// Base provides no-op implementations of Init and Close
// so modules only override what they need.
type Base struct{}

func (b *Base) Init(_ *Deps) error { return nil }
func (b *Base) Close() error       { return nil }
