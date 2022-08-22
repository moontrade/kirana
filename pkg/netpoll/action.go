package netpoll

// Action is an action that occurs after the completion of an event.
type Action int

const (
	// None indicates that no action should occur following an event.
	None Action = iota

	// Close closes the connection.
	Close

	// Shutdown shutdowns the engine.
	Shutdown
)
