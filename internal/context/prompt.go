package context

import (
	"fmt"
	"strings"
)

// MaxPromptChars is the approximate character limit for Whisper's initial prompt.
// 224 tokens ~ 892 characters.
const MaxPromptChars = 892

// DefaultDevTerms are common software development terms that Whisper often
// misrecognizes without context.
var DefaultDevTerms = []string{
	"API", "CLI", "JSON", "YAML", "TOML", "HTTP", "HTTPS", "REST", "gRPC",
	"GraphQL", "WebSocket", "OAuth", "JWT", "UUID", "regex", "stdin", "stdout",
	"stderr", "goroutine", "mutex", "semaphore", "SIGINT", "SIGTERM",
	"GitHub", "GitLab", "CI/CD", "Docker", "Kubernetes", "Terraform",
	"PostgreSQL", "MongoDB", "Redis", "SQLite",
	"TypeScript", "JavaScript", "Python", "Golang", "Rust",
	"refactor", "middleware", "endpoint", "payload", "config",
	"linter", "formatter", "debugger", "profiler",
}

// BuildPrompt constructs a Whisper initial prompt from gathered context.
// The prompt is written as natural-sounding text containing domain terms.
func BuildPrompt(projectName string, terms, skillNames []string) string {
	var parts []string

	if projectName != "" {
		parts = append(parts, fmt.Sprintf("Working on the %s project.", projectName))
	}

	if len(skillNames) > 0 {
		limited := limitSlice(skillNames, 10)
		parts = append(parts, fmt.Sprintf("Available commands: %s.", strings.Join(limited, ", ")))
	}

	if len(terms) > 0 {
		limited := limitSlice(terms, 30)
		parts = append(parts, fmt.Sprintf("Key terms: %s.", strings.Join(limited, ", ")))
	}

	devTermStr := strings.Join(DefaultDevTerms, ", ")
	parts = append(parts, fmt.Sprintf("Common terms: %s.", devTermStr))

	prompt := strings.Join(parts, " ")
	if len(prompt) > MaxPromptChars {
		prompt = prompt[:MaxPromptChars]
		// Trim to last complete word
		if idx := strings.LastIndex(prompt, " "); idx > 0 {
			prompt = prompt[:idx]
		}
	}

	return prompt
}

func limitSlice(s []string, max int) []string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
