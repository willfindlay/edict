package context

import "testing"

func TestWSLToUNC(t *testing.T) {
	tests := []struct {
		distro    string
		linuxPath string
		want      string
	}{
		{"Arch", "/home/will/projects/foo", `\\wsl.localhost\Arch\home\will\projects\foo`},
		{"Ubuntu", "/root", `\\wsl.localhost\Ubuntu\root`},
		{"Arch", "/", `\\wsl.localhost\Arch\`},
	}

	for _, tt := range tests {
		got := WSLToUNC(tt.distro, tt.linuxPath)
		if got != tt.want {
			t.Errorf("WSLToUNC(%q, %q) = %q, want %q", tt.distro, tt.linuxPath, got, tt.want)
		}
	}
}

func TestUNCToLinux(t *testing.T) {
	tests := []struct {
		uncPath string
		want    string
	}{
		{`\\wsl.localhost\Arch\home\will\projects\foo`, "/home/will/projects/foo"},
		{`\\wsl.localhost\Ubuntu\root`, "/root"},
	}

	for _, tt := range tests {
		got := UNCToLinux(tt.uncPath)
		if got != tt.want {
			t.Errorf("UNCToLinux(%q) = %q, want %q", tt.uncPath, got, tt.want)
		}
	}
}
