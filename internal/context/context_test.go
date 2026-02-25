package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testdataDir(t *testing.T) string {
	t.Helper()
	// Navigate from internal/context/ to testdata/
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(wd, "..", "..", "testdata")
}

// --- Path encoding ---

func TestEncodePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/will/projects/foo", "-home-will-projects-foo"},
		{"/", "-"},
		{"/a/b/c", "-a-b-c"},
	}

	for _, tt := range tests {
		got := EncodePath(tt.input)
		if got != tt.want {
			t.Errorf("EncodePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- CLAUDE.md extraction ---

func TestExtractClaudeMDTerms(t *testing.T) {
	dir := testdataDir(t)
	terms := ExtractClaudeMDTerms(dir)

	if len(terms) == 0 {
		t.Fatal("expected non-empty terms from testdata/CLAUDE.md")
	}

	// Check that key backtick terms are extracted
	wantTerms := []string{"goroutine", "RingBuffer", "WhisperClient", "golangci-lint"}
	for _, want := range wantTerms {
		found := false
		for _, term := range terms {
			if term == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected term %q in extracted terms: %v", want, terms)
		}
	}

	// Check that Go declarations are extracted
	wantDecls := []string{"ProcessAudio", "AudioConfig", "DefaultTimeout"}
	for _, want := range wantDecls {
		found := false
		for _, term := range terms {
			if term == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected declaration %q in extracted terms: %v", want, terms)
		}
	}
}

func TestExtractClaudeMDTermsMissing(t *testing.T) {
	terms := ExtractClaudeMDTerms("/nonexistent/path")
	if len(terms) != 0 {
		t.Errorf("expected empty terms for missing dir, got %v", terms)
	}
}

func TestProjectName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/will/projects/edict", "edict"},
		{"/home/will/projects/foo/", "foo"},
		{"/", "."},
	}

	for _, tt := range tests {
		got := ProjectName(tt.input)
		if got != tt.want {
			t.Errorf("ProjectName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- Memory extraction ---

func TestExtractTermsFromMarkdown(t *testing.T) {
	content := "Use `RingBuffer` and `malgo` for audio. The ContextScanner runs periodically."
	terms := extractTermsFromMarkdown(content)

	has := func(want string) bool {
		for _, term := range terms {
			if term == want {
				return true
			}
		}
		return false
	}

	if !has("RingBuffer") {
		t.Errorf("expected RingBuffer in terms: %v", terms)
	}
	if !has("malgo") {
		t.Errorf("expected malgo in terms: %v", terms)
	}
	if !has("ContextScanner") {
		t.Errorf("expected ContextScanner in terms: %v", terms)
	}
}

// --- Skills discovery ---

func TestDiscoverSkillsFromTestdata(t *testing.T) {
	dir := testdataDir(t)

	// Create a fake project structure with .claude/skills pointing to testdata
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(filepath.Join(claudeDir, "skills"), 0o755)
	os.MkdirAll(filepath.Join(claudeDir, "commands"), 0o755)

	// Copy skill dirs
	for _, skill := range []string{"commit", "review"} {
		skillDir := filepath.Join(claudeDir, "skills", skill)
		os.MkdirAll(skillDir, 0o755)
		os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# "+skill), 0o644)
	}

	// Copy command files
	os.WriteFile(filepath.Join(claudeDir, "commands", "deploy.md"), []byte("# deploy"), 0o644)

	// Use a non-existent home to avoid global skills interference
	names := DiscoverSkills(tmpDir, "/nonexistent-home")

	_ = dir // Used for testdata reference

	hasName := func(want string) bool {
		for _, n := range names {
			if n == want {
				return true
			}
		}
		return false
	}

	if !hasName("commit") {
		t.Errorf("expected 'commit' skill, got %v", names)
	}
	if !hasName("review") {
		t.Errorf("expected 'review' skill, got %v", names)
	}
	if !hasName("deploy") {
		t.Errorf("expected 'deploy' command, got %v", names)
	}
}

// --- Prompt construction ---

func TestBuildPrompt(t *testing.T) {
	prompt := BuildPrompt("edict", []string{"RingBuffer", "WhisperClient"}, []string{"commit", "review"})

	if !strings.Contains(prompt, "edict") {
		t.Error("prompt should contain project name")
	}
	if !strings.Contains(prompt, "commit") {
		t.Error("prompt should contain skill names")
	}
	if !strings.Contains(prompt, "RingBuffer") {
		t.Error("prompt should contain terms")
	}
	if !strings.Contains(prompt, "API") {
		t.Error("prompt should contain default dev terms")
	}
}

func TestBuildPromptLengthCap(t *testing.T) {
	// Generate a lot of terms
	var terms []string
	for i := range 200 {
		terms = append(terms, strings.Repeat("x", 10)+string(rune('A'+i%26)))
	}

	prompt := BuildPrompt("bigproject", terms, nil)
	if len(prompt) > MaxPromptChars {
		t.Errorf("prompt exceeds max length: %d > %d", len(prompt), MaxPromptChars)
	}
}

func TestBuildPromptEmpty(t *testing.T) {
	prompt := BuildPrompt("", nil, nil)
	if !strings.Contains(prompt, "Common terms:") {
		t.Error("even empty prompt should have default dev terms")
	}
}
