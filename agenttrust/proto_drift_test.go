package agenttrust

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMoqiProtoMatchesTypeDuckWindowsFrontend(t *testing.T) {
	root := repoRoot(t)
	localPath := filepath.Join(root, "proto", "moqi.proto")
	local, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("read local proto: %v", err)
	}

	refPath := filepath.Join(root, "proto", "FRONTEND_PROTO_REF")
	refData, err := os.ReadFile(refPath)
	if err != nil {
		t.Fatalf("read FRONTEND_PROTO_REF: %v", err)
	}
	ref := ""
	for _, line := range strings.Split(string(refData), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ref = line
		break
	}
	if ref == "" {
		t.Fatal("FRONTEND_PROTO_REF must contain a frontend commit SHA")
	}

	url := "https://raw.githubusercontent.com/KLHA829/TypeDuck-Windows/" + ref + "/proto/moqi.proto"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("fetch frontend proto: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("fetch frontend proto: HTTP %d for %s", resp.StatusCode, url)
	}
	remote, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read frontend proto body: %v", err)
	}

	if string(local) != string(remote) {
		t.Fatalf("proto/moqi.proto drifted from TypeDuck-Windows@%s; copy frontend proto/moqi.proto and regenerate bindings", ref)
	}
}

func TestMoqiProtoIncludesPhase4Markers(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "proto", "moqi.proto"))
	if err != nil {
		t.Fatalf("read proto: %v", err)
	}
	body := string(data)
	for _, marker := range []string{
		"reserved 19;",
		"METHOD_TYPEDUCK_SETTINGS_SNAPSHOT = 100;",
		"METHOD_TYPEDUCK_RECONFIGURE = 103;",
		"METHOD_TYPEDUCK_HEALTH = 104;",
		"string raw_lookup_comment = 6;",
		"message TypeDuckSettingsSnapshot",
		"reserved 26, 100;",
	} {
		if !strings.Contains(body, marker) {
			t.Fatalf("proto/moqi.proto missing Phase 4 marker %q", marker)
		}
	}
	if strings.Contains(body, "METHOD_CLOUD_CLIPBOARD_UPLOAD") {
		t.Fatal("proto/moqi.proto must not define removed METHOD_CLOUD_CLIPBOARD_UPLOAD")
	}
}
