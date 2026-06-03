// Package output formats CLI success output (banners, QR codes, clipboard hints).
package output

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/skip2/go-qrcode"
)

// ReadyOpts configures the post-up success display.
type ReadyOpts struct {
	URL       string
	AppPort   int
	Detached  bool
	Protected bool
	AuthUser  string
	Password  string // shown once when set
	TTL       string
	ExpiresIn string
}

// PrintReady shows the staging URL, optional auth credentials, QR, and next steps.
func PrintReady(o ReadyOpts) {
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	dim := color.New(color.FgHiBlack)

	fmt.Println()
	green.Println("  ┌─────────────────────────────────────────────────────────┐")
	green.Println("  │  Your staging URL is live                               │")
	green.Println("  └─────────────────────────────────────────────────────────┘")
	fmt.Println()
	cyan.Printf("  %s\n", o.URL)
	fmt.Println()

	if o.Protected {
		user := o.AuthUser
		if user == "" {
			user = "solo"
		}
		color.New(color.FgYellow, color.Bold).Printf("  Protected — share with credentials:\n")
		fmt.Printf("    user:     %s\n", user)
		if o.Password != "" {
			fmt.Printf("    password: %s\n", o.Password)
		}
		fmt.Println()
	}

	printQR(o.URL)

	dim.Printf("  Local app port: %d\n", o.AppPort)
	if o.TTL != "" {
		dim.Printf("  Auto teardown:  %s", o.TTL)
		if o.ExpiresIn != "" {
			dim.Printf(" (in %s)", o.ExpiresIn)
		}
		fmt.Println()
	}
	fmt.Println()

	if o.Detached {
		info("Running in the background. Commands:")
		dim.Println("    soloenv status   — check URL and health")
		dim.Println("    soloenv open     — open in browser")
		dim.Println("    soloenv logs     — stream app logs")
		dim.Println("    soloenv down     — stop everything")
	} else {
		info("Press Ctrl+C to stop and tear everything down.")
		dim.Println("    soloenv open  — open URL in browser")
	}

	if copied := tryClipboard(o.URL); copied {
		dim.Println("  (URL copied to clipboard)")
	}
	fmt.Println()
}

func printQR(url string) {
	const size = 6
	code, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return
	}
	bitmap := code.Bitmap()
	color.New(color.FgHiBlack).Println("  Scan on your phone:")
	for y := 0; y < len(bitmap); y += size {
		fmt.Print("  ")
		for x := 0; x < len(bitmap[y]); x += size {
			on := false
			for dy := 0; dy < size && y+dy < len(bitmap); dy++ {
				for dx := 0; dx < size && x+dx < len(bitmap[y]); dx++ {
					if bitmap[y+dy][x+dx] {
						on = true
					}
				}
			}
			if on {
				fmt.Print("██")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func info(s string) {
	color.New(color.FgGreen).Println("  " + s)
}

func tryClipboard(text string) bool {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "clip")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run() == nil
	case "darwin":
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run() == nil
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd := exec.Command("xclip", "-selection", "clipboard")
			cmd.Stdin = strings.NewReader(text)
			return cmd.Run() == nil
		}
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd := exec.Command("wl-copy")
			cmd.Stdin = strings.NewReader(text)
			return cmd.Run() == nil
		}
	}
	return false
}

// Banner prints the SoloEnv wordmark at startup.
func Banner() {
	color.New(color.FgGreen, color.Bold).Println("soloenv")
	color.New(color.FgHiBlack).Println("ephemeral staging for solo devs")
	fmt.Println()
}
