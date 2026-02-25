package context

import (
	"fmt"
	"os/exec"
	"strings"
)

// WSLToUNC converts a Linux path inside a WSL distro to a Windows UNC path.
// Example: WSLToUNC("Arch", "/home/will/projects/foo") returns
// "\\\\wsl.localhost\\Arch\\home\\will\\projects\\foo"
func WSLToUNC(distro, linuxPath string) string {
	// Convert forward slashes to backslashes and prepend UNC root.
	winPath := strings.ReplaceAll(linuxPath, "/", `\`)
	return fmt.Sprintf(`\\wsl.localhost\%s%s`, distro, winPath)
}

// WSLHome returns the home directory path for the default user in the given
// WSL distro. It shells into WSL and reads $HOME.
func WSLHome(distro string) (string, error) {
	cmd := exec.Command("wsl.exe", "-d", distro, "--", "sh", "-c", "echo $HOME")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("wsl home detection: %w", err)
	}
	home := strings.TrimSpace(string(out))
	if home == "" {
		return "", fmt.Errorf("wsl home detection: empty $HOME")
	}
	return home, nil
}

// WSLHomeUNC returns the WSL user home as a UNC path accessible from Windows.
// If wslHome is provided, it's used directly; otherwise it's auto-detected.
func WSLHomeUNC(distro, wslHome string) (string, error) {
	if wslHome == "" {
		var err error
		wslHome, err = WSLHome(distro)
		if err != nil {
			return "", err
		}
	}
	return WSLToUNC(distro, wslHome), nil
}

// UNCToLinux converts a WSL UNC path back to a Linux path.
// Example: UNCToLinux("\\\\wsl.localhost\\Arch\\home\\will") returns "/home/will"
func UNCToLinux(uncPath string) string {
	// Normalize backslashes to forward slashes (filepath.ToSlash only works
	// on Windows, so do it explicitly).
	p := strings.ReplaceAll(uncPath, `\`, "/")
	// Strip //wsl.localhost/<distro> prefix
	const prefix = "//wsl.localhost/"
	if strings.HasPrefix(p, prefix) {
		rest := p[len(prefix):]
		// Skip the distro name
		if idx := strings.Index(rest, "/"); idx >= 0 {
			return rest[idx:]
		}
	}
	return p
}
