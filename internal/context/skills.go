package context

import (
	"os"
	"path/filepath"
	"strings"
)

// DiscoverSkills finds skill and command names from Claude Code skill directories.
func DiscoverSkills(projectDir string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var names []string
	seen := make(map[string]bool)

	// Global skills: ~/.claude/skills/*/SKILL.md
	globalSkills := filepath.Join(home, ".claude", "skills")
	names = append(names, discoverSkillNames(globalSkills, seen)...)

	// Project skills: <project>/.claude/skills/*/SKILL.md
	projectSkills := filepath.Join(projectDir, ".claude", "skills")
	names = append(names, discoverSkillNames(projectSkills, seen)...)

	// Project commands: <project>/.claude/commands/*.md
	commandsDir := filepath.Join(projectDir, ".claude", "commands")
	names = append(names, discoverCommandNames(commandsDir, seen)...)

	return names
}

func discoverSkillNames(dir string, seen map[string]bool) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if seen[name] {
			continue
		}

		// Verify SKILL.md exists
		skillFile := filepath.Join(dir, name, "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}

func discoverCommandNames(dir string, seen map[string]bool) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		cmdName := strings.TrimSuffix(name, ".md")
		if !seen[cmdName] {
			seen[cmdName] = true
			names = append(names, cmdName)
		}
	}
	return names
}
