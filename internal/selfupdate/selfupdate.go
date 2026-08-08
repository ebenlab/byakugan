// Package selfupdate checks GitHub Releases for a newer byakugan and can
// replace the running binary in place. It talks only to the public GitHub
// API and release downloads over HTTPS, verifies the archive's sha256
// against the release's checksums.txt, and swaps the executable atomically.
package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPI = "https://api.github.com"
	repo       = "ebenlab/byakugan"
	// maxArchive caps release downloads; real archives are a few MB.
	maxArchive = 64 << 20
)

// Updater performs release lookups and upgrades. The zero-value fields are
// filled with production defaults by New; tests override API, GOOS/GOARCH,
// and ExecPath.
type Updater struct {
	API    string
	Client *http.Client
	GOOS   string
	GOARCH string
	// ExecPath is the binary to replace. Empty means the running executable,
	// resolved at upgrade time.
	ExecPath string
}

// New returns an Updater with production defaults.
func New() *Updater {
	return &Updater{
		API:    defaultAPI,
		Client: &http.Client{Timeout: 60 * time.Second},
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
	}
}

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// Notice asynchronously checks for a newer release and calls fn with the new
// tag if one exists. It never blocks, never fires for non-release builds,
// respects BYAKUGAN_NO_UPDATE_CHECK, and swallows all errors — an offline
// start must stay silent.
func Notice(current string, fn func(tag string)) {
	if _, ok := parse(current); !ok {
		return // "dev" and other non-release builds
	}
	if os.Getenv("BYAKUGAN_NO_UPDATE_CHECK") != "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if tag, err := New().Check(ctx, current); err == nil && tag != "" {
			fn(tag)
		}
	}()
}

// Check returns the latest release tag if it is strictly newer than
// current, and "" when current is already up to date.
func (u *Updater) Check(ctx context.Context, current string) (string, error) {
	rel, err := u.latest(ctx)
	if err != nil {
		return "", err
	}
	if IsNewer(current, rel.TagName) {
		return rel.TagName, nil
	}
	return "", nil
}

// Upgrade replaces the executable with the latest release when one is newer
// than current. Progress and outcomes are written to out.
func (u *Updater) Upgrade(ctx context.Context, current string, out io.Writer) error {
	if _, ok := parse(current); !ok {
		fmt.Fprintf(out, "byakugan %s is not a release build; grab a release from https://github.com/%s/releases\n", current, repo)
		return nil
	}
	rel, err := u.latest(ctx)
	if err != nil {
		return fmt.Errorf("checking latest release: %w", err)
	}
	if !IsNewer(current, rel.TagName) {
		fmt.Fprintf(out, "byakugan %s is already the latest release\n", current)
		return nil
	}
	a, ok := u.assetFor(rel)
	if !ok {
		return fmt.Errorf("release %s has no asset for %s/%s", rel.TagName, u.GOOS, u.GOARCH)
	}

	fmt.Fprintf(out, "downloading %s…\n", a.Name)
	archive, err := u.download(ctx, a.URL)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", a.Name, err)
	}
	if err := u.verify(ctx, rel, a.Name, archive); err != nil {
		return err
	}
	bin, err := extract(archive, a.Name)
	if err != nil {
		return fmt.Errorf("extracting %s: %w", a.Name, err)
	}
	if err := u.replaceExecutable(bin); err != nil {
		return fmt.Errorf("installing %s (if byakugan came from a package manager, upgrade it there instead): %w", rel.TagName, err)
	}
	fmt.Fprintf(out, "upgraded %s → %s\n", current, rel.TagName)
	return nil
}

func (u *Updater) latest(ctx context.Context) (*release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.API+"/repos/"+repo+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := u.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API: %s", res.Status)
	}
	var rel release
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func (u *Updater) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := u.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, res.Status)
	}
	return io.ReadAll(io.LimitReader(res.Body, maxArchive))
}

// verify checks the archive's sha256 against the release's checksums.txt.
func (u *Updater) verify(ctx context.Context, rel *release, name string, archive []byte) error {
	var sumsURL string
	for _, a := range rel.Assets {
		if a.Name == "checksums.txt" {
			sumsURL = a.URL
			break
		}
	}
	if sumsURL == "" {
		return fmt.Errorf("release %s has no checksums.txt", rel.TagName)
	}
	sums, err := u.download(ctx, sumsURL)
	if err != nil {
		return fmt.Errorf("downloading checksums.txt: %w", err)
	}
	got := hex.EncodeToString(func() []byte { s := sha256.Sum256(archive); return s[:] }())
	for _, line := range strings.Split(string(sums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == name {
			if fields[0] == got {
				return nil
			}
			return fmt.Errorf("checksum mismatch for %s", name)
		}
	}
	return fmt.Errorf("checksums.txt has no entry for %s", name)
}

// assetFor picks this platform's archive, e.g. byakugan_0.3.0_darwin_arm64.tar.gz.
func (u *Updater) assetFor(rel *release) (asset, bool) {
	want := fmt.Sprintf("byakugan_%s_%s_%s.", strings.TrimPrefix(rel.TagName, "v"), u.GOOS, u.GOARCH)
	for _, a := range rel.Assets {
		if strings.HasPrefix(a.Name, want) {
			return a, true
		}
	}
	return asset{}, false
}

// extract pulls the byakugan binary out of a .tar.gz or .zip archive.
func extract(archive []byte, name string) ([]byte, error) {
	if strings.HasSuffix(name, ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == "byakugan.exe" {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(io.LimitReader(rc, maxArchive))
			}
		}
		return nil, fmt.Errorf("byakugan.exe not found in archive")
	}
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("byakugan not found in archive")
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag == tar.TypeReg && filepath.Base(hdr.Name) == "byakugan" {
			return io.ReadAll(io.LimitReader(tr, maxArchive))
		}
	}
}

// replaceExecutable writes bin next to the target executable and renames it
// into place — atomic on POSIX. Windows cannot rename over a running exe, so
// the old binary is moved aside first.
func (u *Updater) replaceExecutable(bin []byte) error {
	exe := u.ExecPath
	if exe == "" {
		var err error
		if exe, err = os.Executable(); err != nil {
			return err
		}
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
	}
	tmp, err := os.CreateTemp(filepath.Dir(exe), ".byakugan-upgrade-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(bin); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		old := exe + ".old"
		os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			return err
		}
		if err := os.Rename(tmp.Name(), exe); err != nil {
			os.Rename(old, exe) // roll back
			return err
		}
		return nil
	}
	return os.Rename(tmp.Name(), exe)
}

// IsNewer reports whether latest is a strictly newer release version than
// current. Unparsable versions (like "dev") never compare as newer.
func IsNewer(current, latest string) bool {
	c, okc := parse(current)
	l, okl := parse(latest)
	if !okc || !okl {
		return false
	}
	for i := 0; i < 3; i++ {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

// parse reads "v1.2.3" (tolerating a -pre/+build suffix) into three ints.
func parse(v string) ([3]int, bool) {
	v = strings.TrimPrefix(v, "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}
