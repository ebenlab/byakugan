package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIsNewer(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"v0.2.0", "v0.3.0", true},
		{"0.2.0", "v0.2.1", true},
		{"v0.2.0", "v1.0.0", true},
		{"v0.9.0", "v0.10.0", true},
		{"v0.3.0", "v0.3.0", false},
		{"v0.3.1", "v0.3.0", false},
		{"v1.0.0", "v0.9.9", false},
		{"dev", "v9.9.9", false},
		{"v0.2.0", "nightly", false},
		{"v0.2.0", "v0.2.1-rc1", true}, // suffix tolerated, compares 0.2.1
		{"v0.2.0", "v0.2", false},      // malformed: never upgrade
	}
	for _, c := range cases {
		if got := IsNewer(c.current, c.latest); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.current, c.latest, got, c.want)
		}
	}
}

// tarGz builds a tar.gz archive holding one file.
func tarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

// zipArchive builds a zip archive holding one file.
func zipArchive(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(content); err != nil {
		t.Fatal(err)
	}
	zw.Close()
	return buf.Bytes()
}

// fakeGitHub serves a latest-release JSON plus its assets and returns an
// Updater pointed at it.
func fakeGitHub(t *testing.T, tag string, assets map[string][]byte) *Updater {
	t.Helper()
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	names := make([]string, 0, len(assets))
	for name := range assets {
		names = append(names, name)
	}
	mux.HandleFunc("/repos/ebenlab/byakugan/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		var list []string
		for _, name := range names {
			list = append(list, fmt.Sprintf(`{"name":%q,"browser_download_url":"%s/dl/%s"}`, name, ts.URL, name))
		}
		fmt.Fprintf(w, `{"tag_name":%q,"assets":[%s]}`, tag, strings.Join(list, ","))
	})
	mux.HandleFunc("/dl/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/dl/")
		body, ok := assets[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write(body)
	})
	return &Updater{API: ts.URL, Client: ts.Client(), GOOS: "linux", GOARCH: "amd64"}
}

func sum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func TestCheck(t *testing.T) {
	u := fakeGitHub(t, "v0.9.0", nil)
	tag, err := u.Check(context.Background(), "v0.2.0")
	if err != nil || tag != "v0.9.0" {
		t.Fatalf("Check = %q, %v; want v0.9.0", tag, err)
	}
	tag, err = u.Check(context.Background(), "v0.9.0")
	if err != nil || tag != "" {
		t.Fatalf("Check same version = %q, %v; want empty", tag, err)
	}
}

func TestUpgradeTarGz(t *testing.T) {
	newBin := []byte("#!/new-binary v0.9.0\n")
	archive := tarGz(t, "byakugan", newBin)
	u := fakeGitHub(t, "v0.9.0", map[string][]byte{
		"byakugan_0.9.0_linux_amd64.tar.gz": archive,
		"checksums.txt":                     []byte(sum(archive) + "  byakugan_0.9.0_linux_amd64.tar.gz\n"),
	})
	exe := filepath.Join(t.TempDir(), "byakugan")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	u.ExecPath = exe

	var out bytes.Buffer
	if err := u.Upgrade(context.Background(), "v0.2.0", &out); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(exe)
	if !bytes.Equal(got, newBin) {
		t.Errorf("binary not replaced: %q", got)
	}
	// Windows has no Unix exec bit (Go reports 0666 for writable files).
	if info, _ := os.Stat(exe); runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		t.Error("replaced binary is not executable")
	}
	if !strings.Contains(out.String(), "upgraded v0.2.0 → v0.9.0") {
		t.Errorf("output = %q", out.String())
	}
}

func TestUpgradeZipPicksWindowsAsset(t *testing.T) {
	newBin := []byte("MZ new windows binary")
	archive := zipArchive(t, "byakugan.exe", newBin)
	u := fakeGitHub(t, "v0.9.0", map[string][]byte{
		"byakugan_0.9.0_windows_amd64.zip": archive,
		"checksums.txt":                    []byte(sum(archive) + "  byakugan_0.9.0_windows_amd64.zip\n"),
	})
	u.GOOS = "windows"
	exe := filepath.Join(t.TempDir(), "byakugan.exe")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	u.ExecPath = exe

	if err := u.Upgrade(context.Background(), "v0.2.0", &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(exe)
	if !bytes.Equal(got, newBin) {
		t.Errorf("binary not replaced: %q", got)
	}
}

func TestUpgradeChecksumMismatchAborts(t *testing.T) {
	archive := tarGz(t, "byakugan", []byte("evil"))
	u := fakeGitHub(t, "v0.9.0", map[string][]byte{
		"byakugan_0.9.0_linux_amd64.tar.gz": archive,
		"checksums.txt":                     []byte(strings.Repeat("0", 64) + "  byakugan_0.9.0_linux_amd64.tar.gz\n"),
	})
	exe := filepath.Join(t.TempDir(), "byakugan")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	u.ExecPath = exe

	err := u.Upgrade(context.Background(), "v0.2.0", &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("err = %v, want checksum mismatch", err)
	}
	if got, _ := os.ReadFile(exe); string(got) != "old" {
		t.Error("binary was replaced despite bad checksum")
	}
}

func TestUpgradeUpToDateAndDevAreNoops(t *testing.T) {
	u := fakeGitHub(t, "v0.9.0", nil)
	exe := filepath.Join(t.TempDir(), "byakugan")
	os.WriteFile(exe, []byte("old"), 0o755)
	u.ExecPath = exe

	var out bytes.Buffer
	if err := u.Upgrade(context.Background(), "v0.9.0", &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "already the latest") {
		t.Errorf("output = %q", out.String())
	}
	out.Reset()
	if err := u.Upgrade(context.Background(), "dev", &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "not a release build") {
		t.Errorf("output = %q", out.String())
	}
	if got, _ := os.ReadFile(exe); string(got) != "old" {
		t.Error("no-op upgrade touched the binary")
	}
}

func TestNoticeSkipsDevAndOptOut(t *testing.T) {
	// A dev build must not even spawn the check.
	fired := make(chan string, 1)
	Notice("dev", func(tag string) { fired <- tag })
	select {
	case tag := <-fired:
		t.Fatalf("Notice fired %q for dev build", tag)
	default:
	}
	t.Setenv("BYAKUGAN_NO_UPDATE_CHECK", "1")
	Notice("v0.1.0", func(tag string) { fired <- tag })
	select {
	case tag := <-fired:
		t.Fatalf("Notice fired %q despite opt-out", tag)
	default:
	}
}
