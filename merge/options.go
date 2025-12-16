package merge

// Option configures a Merger
type Option func(*Merger)

// WithDebug enables debug output
func WithDebug(debug bool) Option {
	return func(m *Merger) {
		m.debug = debug
	}
}
