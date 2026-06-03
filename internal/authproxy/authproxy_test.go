package authproxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// startUpstream returns a backend server and its port.
func startUpstream(t *testing.T) (*httptest.Server, int) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "hello from upstream")
	}))
	t.Cleanup(srv.Close)
	// URL is like http://127.0.0.1:PORT
	idx := strings.LastIndex(srv.URL, ":")
	port, err := strconv.Atoi(srv.URL[idx+1:])
	if err != nil {
		t.Fatal(err)
	}
	return srv, port
}

func waitReady(t *testing.T, url string) {
	t.Helper()
	for i := 0; i < 50; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("proxy never became ready at %s", url)
}

func TestProxyNoAuth(t *testing.T) {
	_, port := startUpstream(t)
	p, err := Start(port, "", "")
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()

	url := "http://127.0.0.1:" + strconv.Itoa(p.Port)
	waitReady(t, url)

	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "upstream") {
		t.Fatalf("status=%d body=%q", resp.StatusCode, body)
	}
}

func TestProxyBasicAuth(t *testing.T) {
	_, port := startUpstream(t)
	p, err := Start(port, "solo", "s3cret")
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()

	url := "http://127.0.0.1:" + strconv.Itoa(p.Port)

	// wait for the proxy to accept connections (expect 401 without creds)
	for i := 0; i < 50; i++ {
		if resp, err := http.Get(url); err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// no creds -> 401
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("without creds: status=%d, want 401", resp.StatusCode)
	}

	// wrong creds -> 401
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth("solo", "wrong")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong creds: status=%d, want 401", resp.StatusCode)
	}

	// correct creds -> 200
	req, _ = http.NewRequest("GET", url, nil)
	req.SetBasicAuth("solo", "s3cret")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "upstream") {
		t.Fatalf("correct creds: status=%d body=%q", resp.StatusCode, body)
	}
}
