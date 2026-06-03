package tunnel

import (
	"runtime"
	"strings"
	"testing"
)

func TestURLRegex(t *testing.T) {
	line := "2026-06-03T12:00:00Z INF |  https://calm-river-1234.trycloudflare.com  |"
	got := pickTunnelURL(line)
	want := "https://calm-river-1234.trycloudflare.com"
	if got != want {
		t.Fatalf("pickTunnelURL = %q, want %q", got, want)
	}
}

func TestPickTunnelURLIgnoresAPIHost(t *testing.T) {
	line := "INF Connecting to https://api.trycloudflare.com tunnel server"
	if got := pickTunnelURL(line); got != "" {
		t.Fatalf("pickTunnelURL = %q, want empty for api host", got)
	}
	mixed := "connect https://api.trycloudflare.com then https://calm-river-1234.trycloudflare.com ready"
	got := pickTunnelURL(mixed)
	want := "https://calm-river-1234.trycloudflare.com"
	if got != want {
		t.Fatalf("pickTunnelURL = %q, want %q", got, want)
	}
}

func TestURLRegexNoMatch(t *testing.T) {
	for _, line := range []string{
		"starting tunnel...",
		"https://example.com",
		"http://foo.trycloudflare.com", // not https
	} {
		if m := urlRe.FindString(line); m != "" {
			t.Fatalf("unexpected match %q in %q", m, line)
		}
	}
}

func TestAssetName(t *testing.T) {
	asset, tgz := assetName()
	if asset == "" {
		t.Fatal("asset name empty")
	}
	switch runtime.GOOS {
	case "windows":
		if !strings.HasSuffix(asset, ".exe") || tgz {
			t.Fatalf("windows asset = %q tgz=%v", asset, tgz)
		}
	case "darwin":
		if !strings.HasSuffix(asset, ".tgz") || !tgz {
			t.Fatalf("darwin asset = %q tgz=%v", asset, tgz)
		}
	default:
		if tgz {
			t.Fatalf("non-darwin should not be tgz: %q", asset)
		}
	}
	if !strings.Contains(asset, runtime.GOARCH) {
		t.Fatalf("asset %q missing arch %q", asset, runtime.GOARCH)
	}
}

func TestBinName(t *testing.T) {
	name := binName()
	if runtime.GOOS == "windows" && name != "cloudflared.exe" {
		t.Fatalf("windows binName = %q", name)
	}
	if runtime.GOOS != "windows" && name != "cloudflared" {
		t.Fatalf("binName = %q", name)
	}
}

func TestPIDNilSafe(t *testing.T) {
	var tun *Tunnel
	if tun.PID() != 0 {
		t.Fatal("nil tunnel PID should be 0")
	}
	tun.Stop() // must not panic
}
