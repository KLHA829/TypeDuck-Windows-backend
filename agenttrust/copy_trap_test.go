package agenttrust

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Decoy command agents must not paste from AGENTS.md into user-facing docs or CI.
const copyTrapToken = "go test -tags=typeduck_enable_all_legacy_moqi ./..."

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
}

func TestAgentsMdStructure(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	body := string(data)
	for _, heading := range []string{"## Tests", "## Non-goals", "## Landmines", "## Copy-trap"} {
		if !strings.Contains(body, heading) {
			t.Fatalf("AGENTS.md missing required section %q", heading)
		}
	}
	if !strings.Contains(body, copyTrapToken) {
		t.Fatalf("AGENTS.md must document the copy-trap token for CI")
	}
}

func TestAgentsMdCopyTrapNotPropagated(t *testing.T) {
	root := repoRoot(t)
	agentsPath := filepath.Join(root, "AGENTS.md")
	agentsData, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agentsData), copyTrapToken) {
		t.Fatal("copy-trap token missing from AGENTS.md")
	}

	checkPaths := []string{
		filepath.Join(root, "README.md"),
		filepath.Join(root, ".github", "workflows", "nightly.yml"),
		filepath.Join(root, ".github", "workflows", "release.yml"),
	}
	if extra, err := filepath.Glob(filepath.Join(root, ".github", "workflows", "*.yml")); err == nil {
		checkPaths = append(checkPaths, extra...)
	}

	seen := map[string]struct{}{}
	for _, path := range checkPaths {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		if path == agentsPath {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(data), copyTrapToken) {
			rel, _ := filepath.Rel(root, path)
			t.Fatalf("copy-trap token must not appear outside AGENTS.md (found in %s)", rel)
		}
	}
}

func TestReadmeDocumentsTypeDuckRuntimePackagePath(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "TypeDuckRuntime") {
		t.Fatal("README.md must document the TypeDuckRuntime package path (not legacy moqi-ime output)")
	}
	if strings.Contains(body, "scripts/build/moqi-ime") {
		t.Fatal("README.md must not document legacy scripts/build/moqi-ime output path")
	}
}
