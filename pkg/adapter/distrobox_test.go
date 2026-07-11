package adapter

import (
	"testing"

	"github.com/DaviMGDev/unipm/pkg/config"
)

func TestNewDistroboxAdapter(t *testing.T) {
	cfg := config.DistroboxConfig{
		ContainerName:  "arch-dev",
		PackageManager: "yay",
	}

	a := NewDistroboxAdapter("arch", cfg)

	if a.Name() != "distrobox-arch-dev" {
		t.Errorf("Name() = %q, want %q", a.Name(), "distrobox-arch-dev")
	}
	if a.ContainerName != "arch-dev" {
		t.Errorf("ContainerName = %q, want %q", a.ContainerName, "arch-dev")
	}
	if a.PackageManager != "yay" {
		t.Errorf("PackageManager = %q, want %q", a.PackageManager, "yay")
	}
	if a.Nickname != "arch" {
		t.Errorf("Nickname = %q, want %q", a.Nickname, "arch")
	}
}

func TestDistroboxAdapter_IsAvailable_NotOnPath(t *testing.T) {
	// This test verifies the method doesn't panic.
	// The actual return depends on the environment.
	a := NewDistroboxAdapter("arch", config.DistroboxConfig{
		ContainerName:  "nonexistent-container-xyz",
		PackageManager: "pacman",
	})

	available := a.IsAvailable()
	// Should be false in CI/test environments
	t.Logf("distrobox IsAvailable() = %v", available)
}

func TestBuildSearchArgs(t *testing.T) {
	tests := []struct {
		pm       string
		query    string
		wantArgs []string
	}{
		{
			pm:    "apt",
			query: "htop",
			wantArgs: []string{"apt", "search", "htop"},
		},
		{
			pm:    "pacman",
			query: "htop",
			wantArgs: []string{"pacman", "-Ss", "htop"},
		},
		{
			pm:    "yay",
			query: "htop",
			wantArgs: []string{"yay", "-Ss", "htop"},
		},
		{
			pm:    "dnf",
			query: "htop",
			wantArgs: []string{"dnf", "search", "htop"},
		},
		{
			pm:    "zypper",
			query: "htop",
			wantArgs: []string{"zypper", "search", "htop"},
		},
		{
			pm:       "unknown",
			query:    "htop",
			wantArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.pm, func(t *testing.T) {
			a := NewDistroboxAdapter("test", config.DistroboxConfig{
				ContainerName:  "test-container",
				PackageManager: tt.pm,
			})

			args := a.buildSearchArgs(tt.query)

			if tt.wantArgs == nil {
				if args != nil {
					t.Errorf("expected nil args for unsupported PM, got %v", args)
				}
				return
			}

			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args = %v, want %v", args, tt.wantArgs)
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestBuildInstallArgs(t *testing.T) {
	tests := []struct {
		pm       string
		pkg      string
		wantArgs []string
	}{
		{pm: "apt", pkg: "htop", wantArgs: []string{"sudo", "apt", "install", "-y", "htop"}},
		{pm: "pacman", pkg: "htop", wantArgs: []string{"sudo", "pacman", "-S", "--noconfirm", "htop"}},
		{pm: "yay", pkg: "htop", wantArgs: []string{"yay", "-S", "--noconfirm", "htop"}},
		{pm: "dnf", pkg: "htop", wantArgs: []string{"sudo", "dnf", "install", "-y", "htop"}},
	}

	for _, tt := range tests {
		t.Run(tt.pm, func(t *testing.T) {
			a := NewDistroboxAdapter("test", config.DistroboxConfig{
				ContainerName:  "test-container",
				PackageManager: tt.pm,
			})
			args := a.buildInstallArgs(tt.pkg)
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args = %v, want %v", args, tt.wantArgs)
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestBuildUninstallArgs(t *testing.T) {
	tests := []struct {
		pm       string
		pkg      string
		wantArgs []string
	}{
		{pm: "apt", pkg: "htop", wantArgs: []string{"sudo", "apt", "remove", "-y", "htop"}},
		{pm: "pacman", pkg: "htop", wantArgs: []string{"sudo", "pacman", "-R", "--noconfirm", "htop"}},
		{pm: "yay", pkg: "htop", wantArgs: []string{"yay", "-R", "--noconfirm", "htop"}},
		{pm: "dnf", pkg: "htop", wantArgs: []string{"sudo", "dnf", "remove", "-y", "htop"}},
	}

	for _, tt := range tests {
		t.Run(tt.pm, func(t *testing.T) {
			a := NewDistroboxAdapter("test", config.DistroboxConfig{
				ContainerName:  "test-container",
				PackageManager: tt.pm,
			})
			args := a.buildUninstallArgs(tt.pkg)
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args = %v, want %v", args, tt.wantArgs)
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestParsePMLine(t *testing.T) {
	tests := []struct {
		pm      string
		line    string
		want    string // name
		wantVer string // version (substring match)
	}{
		// apt format
		{pm: "apt", line: "htop/stable 3.4.1-5 amd64", want: "htop", wantVer: "3.4.1-5"},
		{pm: "apt", line: "  interactive viewer", want: ""}, // indented description, skipped
		{pm: "apt", line: "Sorting...", want: ""},            // header, skipped

		// pacman format
		{pm: "pacman", line: "core/htop 3.4.1-1", want: "htop", wantVer: "3.4.1-1"},

		// yay format
		{pm: "yay", line: "aur/htop-git 3.4.1.r12.gabc123-1 (42)", want: "htop-git", wantVer: "3.4.1.r12.gabc123-1"},

		// dnf format
		{pm: "dnf", line: "htop.x86_64  3.4.1-1.fc40  fedora", want: "htop", wantVer: "3.4.1-1.fc40"},
	}

	for _, tt := range tests {
		t.Run(tt.pm+"/"+tt.line, func(t *testing.T) {
			name, version, _ := parsePMLine(tt.line, tt.pm)
			if name != tt.want {
				t.Errorf("name = %q, want %q", name, tt.want)
			}
			if tt.wantVer != "" && version != tt.wantVer {
				t.Errorf("version = %q, want %q", version, tt.wantVer)
			}
		})
	}
}

func TestParseDistroboxSearch_Empty(t *testing.T) {
	pkgs := parseDistroboxSearch("", "test-container", "apt")
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages from empty output, got %d", len(pkgs))
	}
}

func TestParseDistroboxInfo_Empty(t *testing.T) {
	d := parseDistroboxInfo("", "apt")
	if d.Name != "" {
		t.Errorf("expected empty Details from empty output, got %+v", d)
	}
}

func TestDistroboxAdapter_IntegrationSkip(t *testing.T) {
	a := NewDistroboxAdapter("test", config.DistroboxConfig{
		ContainerName:  "nonexistent-xyz-container",
		PackageManager: "pacman",
	})

	if a.IsAvailable() {
		t.Skip("distrobox with test container available — unexpected in CI")
	}
	t.Log("distrobox correctly unavailable for nonexistent container")
}
