// Package tunnel manages a free Cloudflare quick tunnel: it ensures the
// cloudflared binary is available (downloading it on first use), runs it against
// a local port, and extracts the public *.trycloudflare.com URL.
package tunnel

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const releaseBase = "https://github.com/cloudflare/cloudflared/releases/latest/download"

var urlRe = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)

// cloudflared logs its control-plane host (api.trycloudflare.com) before the
// random quick-tunnel URL — ignore reserved subdomains.
var blockedTunnelHosts = map[string]bool{
	"api.trycloudflare.com": true,
	"www.trycloudflare.com": true,
}

func pickTunnelURL(line string) string {
	for _, m := range urlRe.FindAllString(line, -1) {
		u, err := url.Parse(m)
		if err != nil {
			continue
		}
		host := strings.ToLower(u.Hostname())
		if blockedTunnelHosts[host] {
			continue
		}
		return m
	}
	return ""
}

// Tunnel is a running cloudflared process and its public URL.
type Tunnel struct {
	URL string
	cmd *exec.Cmd
}

// PID returns the cloudflared process id, or 0 if not running.
func (t *Tunnel) PID() int {
	if t == nil || t.cmd == nil || t.cmd.Process == nil {
		return 0
	}
	return t.cmd.Process.Pid
}

// Stop terminates the cloudflared process.
func (t *Tunnel) Stop() {
	if t == nil || t.cmd == nil || t.cmd.Process == nil {
		return
	}
	_ = t.cmd.Process.Kill()
	_ = t.cmd.Wait()
}

// Ensure returns a path to a usable cloudflared binary, downloading it into the
// user cache directory if it is not already on PATH or cached.
func Ensure() (string, error) {
	if p, err := exec.LookPath("cloudflared"); err == nil {
		return p, nil
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cacheDir, "soloenv")
	dest := filepath.Join(dir, binName())
	if fileExists(dest) {
		return dest, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := download(dest); err != nil {
		return "", fmt.Errorf("could not download cloudflared: %w", err)
	}
	return dest, nil
}

// Start launches cloudflared against http://127.0.0.1:<port> and waits for the URL.
func Start(bin string, port int, timeout time.Duration) (*Tunnel, error) {
	return StartURL(bin, fmt.Sprintf("http://127.0.0.1:%d", port), timeout)
}

// StartURL runs cloudflared tunnel --url <targetURL>.
func StartURL(bin, targetURL string, timeout time.Duration) (*Tunnel, error) {
	cmd := exec.Command(bin, "tunnel", "--no-autoupdate", "--url", targetURL)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	urlCh := make(chan string, 1)
	scan := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			if m := pickTunnelURL(scanner.Text()); m != "" {
				select {
				case urlCh <- m:
				default:
				}
			}
		}
	}
	go scan(stdout)
	go scan(stderr)

	select {
	case u := <-urlCh:
		return &Tunnel{URL: u, cmd: cmd}, nil
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("timed out after %s waiting for a tunnel URL", timeout)
	}
}

func download(dest string) error {
	asset, tgz := assetName()
	url := releaseBase + "/" + asset
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s for %s", resp.Status, url)
	}

	if tgz {
		return extractFromTarGz(resp.Body, "cloudflared", dest)
	}
	return writeExecutable(resp.Body, dest)
}

func writeExecutable(r io.Reader, dest string) error {
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	return nil
}

// extractFromTarGz pulls a single named entry out of a .tgz stream (macOS asset).
func extractFromTarGz(r io.Reader, name, dest string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("%q not found in archive", name)
		}
		if err != nil {
			return err
		}
		if filepath.Base(hdr.Name) == name {
			return writeExecutable(tr, dest)
		}
	}
}

// assetName returns the cloudflared release asset for this OS/arch and whether
// it is a .tgz archive (true only for macOS).
func assetName() (string, bool) {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("cloudflared-windows-%s.exe", runtime.GOARCH), false
	case "darwin":
		return fmt.Sprintf("cloudflared-darwin-%s.tgz", runtime.GOARCH), true
	default: // linux and others ship a raw binary
		return fmt.Sprintf("cloudflared-%s-%s", runtime.GOOS, runtime.GOARCH), false
	}
}

func binName() string {
	if runtime.GOOS == "windows" {
		return "cloudflared.exe"
	}
	return "cloudflared"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
