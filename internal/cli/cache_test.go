package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicWritesCompleteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache", "export.json")

	if err := writeFileAtomic(path, []byte(`{"old":true}`), 0o600); err != nil {
		t.Fatalf("writeFileAtomic() initial write error = %v", err)
	}
	if err := writeFileAtomic(path, []byte(`{"new":true}`), 0o600); err != nil {
		t.Fatalf("writeFileAtomic() replace error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got, want := string(data), `{"new":true}`; got != want {
		t.Fatalf("cache content = %q, want %q", got, want)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("cache mode = %v, want %v", got, want)
	}
}

func TestWriteFileAtomicRemovesTempFileAfterRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "export.json")

	if err := writeFileAtomic(path, []byte(`[]`), 0o600); err != nil {
		t.Fatalf("writeFileAtomic() error = %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(dir, ".export.json.*.tmp"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("temp files left behind: %v", matches)
	}
}
