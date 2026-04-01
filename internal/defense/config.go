// Package defense provides scanner defense to block malicious IPs at the TCP connection level.
package defense

import (
	"time"
)

// Config holds configuration for the scanner defense system.
type Config struct {
	// Enabled controls whether scanner defense is active.
	Enabled bool `json:"enabled"`

	// RateLimit is the maximum number of requests per RateWindow before blocking.
	RateLimit int `json:"rate_limit"`

	// RateWindow is the time window for rate limiting.
	RateWindow time.Duration `json:"rate_window"`

	// ErrorRateThreshold is the error rate (0.0-1.0) above which an IP is blocked.
	// For example, 0.9 means 90% error rate triggers a block.
	ErrorRateThreshold float64 `json:"error_rate_threshold"`

	// MinRequestsForAnalysis is the minimum number of requests needed before
	// analyzing error rates. This prevents blocking after just one or two errors.
	MinRequestsForAnalysis int `json:"min_requests"`

	// SuspiciousPathThreshold is the number of suspicious path hits that trigger a block.
	SuspiciousPathThreshold int `json:"suspicious_path_threshold"`

	// BlockDuration is how long an IP remains blocked.
	BlockDuration time.Duration `json:"block_duration"`

	// Whitelist contains CIDR notation ranges that should never be blocked.
	Whitelist []string `json:"whitelist"`

	// BlockCommand is an optional external command to run when an IP is blocked.
	// The placeholder {ip} is replaced with the blocked IP address.
	// Example: "pfctl -t mitto_blocked -T add {ip}" or "iptables -A INPUT -s {ip} -j DROP"
	// If empty, no external command is executed (only in-memory blocklist is used).
	// The command is executed asynchronously so it doesn't block request handling.
	BlockCommand string `json:"block_command"`

	// PersistPath is the path to the blocklist persistence file.
	// If empty, persistence is disabled.
	PersistPath string `json:"persist_path"`
}

// DefaultConfig returns sensible default configuration for scanner defense.
func DefaultConfig() Config {
	return Config{
		Enabled:                 false, // Disabled by default - user must opt-in
		RateLimit:               100,
		RateWindow:              time.Minute,
		ErrorRateThreshold:      0.9,
		MinRequestsForAnalysis:  10,
		SuspiciousPathThreshold: 5,
		BlockDuration:           7 * 24 * time.Hour, // 7 days
		Whitelist:               []string{"127.0.0.0/8", "::1/128"},
	}
}
