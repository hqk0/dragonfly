package server

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadResourcesIgnoresHiddenFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"._01-hqlt-Res", ".DS_Store"} {
		if err := os.WriteFile(
			filepath.Join(dir, name), []byte("not a resource pack"), 0o600,
		); err != nil {
			t.Fatal(err)
		}
	}
	packs, err := loadResources(dir, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 0 {
		t.Fatalf("loaded packs = %d, want none", len(packs))
	}
}

func TestLoadResourcesWithKeyFile(t *testing.T) {
	dir := t.TempDir()
	packDir := filepath.Join(dir, "01-test-pack")
	if err := os.MkdirAll(packDir, 0o755); err != nil {
		t.Fatal(err)
	}

	manifestJSON := `{
		"format_version": 2,
		"header": {
			"name": "Test Pack",
			"description": "Test Pack Description",
			"uuid": "f81d4fae-7dec-11d0-a765-00a0c91e6bf6",
			"version": [1, 0, 0],
			"min_engine_version": [1, 20, 0]
		},
		"modules": [
			{
				"type": "resources",
				"uuid": "f81d4fae-7dec-11d0-a765-00a0c91e6bf7",
				"version": [1, 0, 0]
			}
		]
	}`
	if err := os.WriteFile(filepath.Join(packDir, "manifest.json"), []byte(manifestJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	keyContent := "secret_encryption_key_12345\n"
	if err := os.WriteFile(filepath.Join(dir, "01-test-pack.key"), []byte(keyContent), 0o644); err != nil {
		t.Fatal(err)
	}

	packs, err := loadResources(dir, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 1 {
		t.Fatalf("loaded packs count = %d, want 1", len(packs))
	}
	if key := packs[0].ContentKey(); key != "secret_encryption_key_12345" {
		t.Fatalf("got content key %q, want %q", key, "secret_encryption_key_12345")
	}
	if !packs[0].Encrypted() {
		t.Fatalf("expected pack to be marked as encrypted")
	}
}

func TestLoadResourcesFromCDN(t *testing.T) {
	dir := t.TempDir()
	packDir := filepath.Join(dir, "01-test-pack")
	if err := os.MkdirAll(packDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := []byte(`{
		"format_version": 2,
		"header": {
			"name": "CDN Test Pack",
			"description": "CDN Test Pack",
			"uuid": "f81d4fae-7dec-11d0-a765-00a0c91e6bf6",
			"version": [1, 2, 3],
			"min_engine_version": [1, 20, 0]
		},
		"modules": [{
			"type": "resources",
			"uuid": "f81d4fae-7dec-11d0-a765-00a0c91e6bf7",
			"version": [1, 2, 3]
		}]
	}`)
	if err := os.WriteFile(filepath.Join(packDir, "manifest.json"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "01-test-pack.key"), []byte("test-content-key"), 0o600); err != nil {
		t.Fatal(err)
	}

	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	w, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(manifest); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive.Bytes())
	}))
	defer server.Close()

	packs, err := loadResources(dir, map[string]string{
		"f81d4fae-7dec-11d0-a765-00a0c91e6bf6": server.URL + "/pack.mcpack",
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 1 {
		t.Fatalf("loaded packs count = %d, want 1", len(packs))
	}
	if got := packs[0].DownloadURL(); got != server.URL+"/pack.mcpack" {
		t.Fatalf("download URL = %q", got)
	}
	if got := packs[0].ContentKey(); got != "test-content-key" {
		t.Fatalf("content key = %q", got)
	}
}

func TestLoadResourcesRequiresCDN(t *testing.T) {
	dir := t.TempDir()
	packDir := filepath.Join(dir, "01-test-pack")
	if err := os.MkdirAll(packDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"format_version": 2,
		"header": {
			"name": "Test Pack",
			"description": "Test Pack",
			"uuid": "f81d4fae-7dec-11d0-a765-00a0c91e6bf6",
			"version": [1, 0, 0],
			"min_engine_version": [1, 20, 0]
		},
		"modules": [{
			"type": "resources",
			"uuid": "f81d4fae-7dec-11d0-a765-00a0c91e6bf7",
			"version": [1, 0, 0]
		}]
	}`
	if err := os.WriteFile(filepath.Join(packDir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadResources(dir, nil, true); err == nil {
		t.Fatal("expected missing CDN URL error")
	}
}

func TestRequiredResourcesRejectMissingOrEmptyDirectory(t *testing.T) {
	for _, test := range []struct {
		name string
		dir  func(t *testing.T) string
	}{
		{name: "missing", dir: func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing") }},
		{name: "empty", dir: func(t *testing.T) string { return t.TempDir() }},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := DefaultConfig()
			config.World.SaveData = false
			config.Players.SaveData = false
			config.Resources.Folder = test.dir(t)
			config.Resources.Required = true
			if _, err := config.Config(nil); err == nil {
				t.Fatal("expected required resource directory error")
			}
		})
	}
}
