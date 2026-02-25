package context

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// backtickTerm matches backtick-quoted terms (code references).
	backtickTerm = regexp.MustCompile("`([A-Za-z][A-Za-z0-9_./-]+)`")

	// camelCase matches CamelCase or camelCase identifiers.
	camelCase = regexp.MustCompile(`\b([A-Z][a-z]+(?:[A-Z][a-z]+)+|[a-z]+(?:[A-Z][a-z]+)+)\b`)

	// goDecl matches Go-style declarations (func, type, var, const).
	goDecl = regexp.MustCompile(`(?:func|type|var|const)\s+([A-Za-z_][A-Za-z0-9_]*)`)
)

// ExtractClaudeMDTerms reads CLAUDE.md files from a project and extracts
// domain-specific terms for the Whisper prompt.
func ExtractClaudeMDTerms(projectDir string) []string {
	candidates := []string{
		filepath.Join(projectDir, "CLAUDE.md"),
		filepath.Join(projectDir, ".claude", "CLAUDE.md"),
	}

	seen := make(map[string]bool)
	var terms []string

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)

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

		for _, match := range goDecl.FindAllStringSubmatch(content, -1) {
			term := match[1]
			if !seen[term] && len(term) >= 3 {
				seen[term] = true
				terms = append(terms, term)
			}
		}
	}

	return terms
}

// ProjectName extracts the project name from a directory path.
func ProjectName(projectDir string) string {
	return filepath.Base(strings.TrimRight(projectDir, "/"))
}
