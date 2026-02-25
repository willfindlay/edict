package context

import (
	"os"
	"path/filepath"
	"strings"
)

// EncodePath converts a filesystem path to Claude Code's encoded format.
// /home/will/projects/foo becomes -home-will-projects-foo
func EncodePath(path string) string {
	return strings.ReplaceAll(path, "/", "-")
}

// ExtractMemoryTerms reads the project's auto-memory MEMORY.md and extracts terms.
// canonicalProjectDir is always the Linux-format path (for EncodePath compatibility).
// homeDir is the OS-accessible home directory (UNC on Windows, native on Linux).
func ExtractMemoryTerms(projectDir, homeDir, canonicalProjectDir string) []string {
	encoded := EncodePath(canonicalProjectDir)

	memoryPath := filepath.Join(homeDir, ".claude", "projects", encoded, "memory", "MEMORY.md")
	data, err := os.ReadFile(memoryPath)
	if err != nil {
		return nil
	}

	return extractTermsFromMarkdown(string(data))
}

func extractTermsFromMarkdown(content string) []string {
	seen := make(map[string]bool)
	var terms []string

	for _, match := range backtickTerm.FindAllStringSubmatch(content, -1) {
		term := match[1]
		if !seen[term] && len(term) >= 3 {
			seen[term] = true
			terms = append(terms, term)
		}
	}

	for _, match := range camelCase.FindAllStringSubmatch(content, -1) {
		term := match[1]
		if !seen[term] {
			seen[term] = true
			terms = append(terms, term)
		}
	}

	return terms
}
