package context

// ClaudeProcess represents a detected Claude Code process.
type ClaudeProcess struct {
	PID int
	CWD string // OS-accessible working directory (UNC path on Windows, native on Linux)

	// CanonicalCWD is the Linux-format path, needed for EncodePath() to match
	// Claude Code's memory directory encoding scheme. On Linux this equals CWD.
	CanonicalCWD string
}
