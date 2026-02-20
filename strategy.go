package erl

// Strategy defines the behavior when a rate limit is reached.
type Strategy int

const (
	// Block returns ErrLimitExceeded immediately when the limit is hit.
	Block Strategy = iota
	// BlockWithQueue blocks by default but exposes a Wait method on the error
	// so callers can opt into waiting until the window resets.
	BlockWithQueue
	// LogOnly lets the request through and calls the OnLimitReached callback.
	LogOnly
)

func (s Strategy) String() string {
	switch s {
	case Block:
		return "Block"
	case BlockWithQueue:
		return "BlockWithQueue"
	case LogOnly:
		return "LogOnly"
	default:
		return "Unknown"
	}
}
