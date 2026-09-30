package engine

// State represents the current condition of the network and gateway session.
type State int

const (
	// StateOffline indicates no network connectivity or gateway unreachable.
	StateOffline State = iota

	// StateCaptive indicates traffic is intercepted by FortiGate; login required.
	StateCaptive

	// StateOnline indicates unblocked internet access; keepalive active.
	StateOnline

	// StateCooldown indicates consecutive failures; logins paused to prevent lockout.
	StateCooldown

	// StateSecurityHalted indicates a security violation (certificate mismatch, foreign portal, or untrusted redirect). Polling is suspended.
	StateSecurityHalted
)

// String returns a human-readable label for logs and UI status.
func (s State) String() string {
	switch s {
	case StateOffline:
		return "Offline"
	case StateCaptive:
		return "Captive Portal Detected"
	case StateOnline:
		return "Online"
	case StateCooldown:
		return "Cooldown (Suspended)"
	case StateSecurityHalted:
		return "Security Halted (Manual Verification Required)"
	default:
		return "Unknown"
	}
}
