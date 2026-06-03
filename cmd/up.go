package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fleames/soloenv-cli/internal/authproxy"
	"github.com/fleames/soloenv-cli/internal/docker"
	"github.com/fleames/soloenv-cli/internal/output"
	"github.com/fleames/soloenv-cli/internal/preflight"
	"github.com/fleames/soloenv-cli/internal/project"
	"github.com/fleames/soloenv-cli/internal/state"
	"github.com/fleames/soloenv-cli/internal/tunnel"
	"github.com/spf13/cobra"
)

var (
	flagPort     int
	flagService  string
	flagNoBuild  bool
	flagDetach   bool
	flagPassword string
	flagProtect  bool
	flagTTL      string
	flagEnvFile  string
	flagOpen     bool
)

const tunnelTimeout = 60 * time.Second

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start your app and get a public staging URL",
	Long: `Builds and runs your Docker app, then opens a Cloudflare quick tunnel.

By default stays in the foreground (Ctrl+C tears down). Use --detach to return
to your shell while the environment keeps running.

Examples:
  soloenv up
  soloenv up --detach --ttl 4h
  soloenv up --protect
  soloenv up --password demo123 --env-file .env.staging`,
	RunE: runUp,
}

func init() {
	upCmd.Flags().IntVarP(&flagPort, "port", "p", 0, "host port to expose (overrides auto-detect)")
	upCmd.Flags().StringVarP(&flagService, "service", "s", "", "compose service when several publish ports")
	upCmd.Flags().BoolVar(&flagNoBuild, "no-build", false, "skip image build")
	upCmd.Flags().BoolVarP(&flagDetach, "detach", "d", false, "run in background; use soloenv down to stop")
	upCmd.Flags().StringVar(&flagPassword, "password", "", "HTTP basic auth password for the public URL")
	upCmd.Flags().BoolVar(&flagProtect, "protect", false, "enable basic auth with an auto-generated password")
	upCmd.Flags().StringVar(&flagTTL, "ttl", "", "auto teardown after duration (e.g. 2h, 45m)")
	upCmd.Flags().StringVar(&flagEnvFile, "env-file", "", "env file for docker (default: .env.staging or .env.soloenv)")
	upCmd.Flags().BoolVar(&flagOpen, "open", false, "open the staging URL in your browser when ready")
}

func runUp(cmd *cobra.Command, args []string) error {
	output.Banner()

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if state.Exists(absDir) {
		if st, err := state.Load(absDir); err == nil {
			return fmt.Errorf("environment already running: %s\n  run `soloenv down` first", st.URL)
		}
		return fmt.Errorf("environment already running; run `soloenv down` first")
	}

	if err := preflight.CheckDocker(); err != nil {
		return err
	}

	proj, err := project.Detect(absDir)
	if err != nil {
		return err
	}
	cfg, err := project.LoadConfig(absDir)
	if err != nil {
		return err
	}

	envFile, err := project.ResolveEnvFile(absDir, cfg, flagEnvFile)
	if err != nil {
		return err
	}
	if envFile != "" {
		info("Using env file: %s", filepath.Base(envFile))
	}

	authUser, authPass, err := project.ResolveAuth(cfg, flagPassword, flagProtect)
	if err != nil {
		return err
	}

	ttl, err := project.ResolveTTL(cfg, flagTTL)
	if err != nil {
		return fmt.Errorf("invalid ttl: %w", err)
	}

	port := flagPort
	if port == 0 {
		port = cfg.Port
	}
	service := flagService
	if service == "" {
		service = cfg.Service
	}
	build := resolveBuild(cfg)

	st := &state.State{
		ProjectDir: absDir,
		StartedAt:  time.Now(),
		Detached:   flagDetach,
	}

	switch proj.Kind {
	case project.KindCompose:
		resolved, err := project.ResolveComposePort(absDir, proj.ComposeFile, service, port)
		if err != nil {
			return err
		}
		port = resolved
		info("Starting compose project (%s)...", proj.ComposeFile)
		if err := docker.ComposeUp(absDir, proj.ComposeFile, build, envFile); err != nil {
			return fmt.Errorf("failed to start app: %w", err)
		}
		st.Mode = state.ModeCompose
		st.ComposeFile = proj.ComposeFile

	case project.KindDockerfile:
		containerPort := project.DockerfileExpose(absDir, proj.Dockerfile)
		if port == 0 {
			port = containerPort
		}
		if port == 0 {
			return fmt.Errorf("could not determine port â€” add EXPOSE or pass --port")
		}
		if containerPort == 0 {
			containerPort = port
		}
		tag := "soloenv-" + sanitize(filepathBase(absDir))
		if build {
			info("Building image %s...", tag)
			if err := docker.Build(absDir, proj.Dockerfile, tag); err != nil {
				return fmt.Errorf("failed to build: %w", err)
			}
		}
		info("Starting container (%d -> %d)...", port, containerPort)
		id, err := docker.Run(absDir, tag, port, containerPort, envFile)
		if err != nil {
			return err
		}
		st.Mode = state.ModeDockerfile
		st.ContainerID = id
		st.ImageTag = tag
	}
	st.AppPort = port

	tunnelPort := port
	var localProxy *authproxy.Server
	displayPassword := ""

	if authPass != "" {
		st.AuthEnabled = true
		st.AuthUser = authUser
		displayPassword = authPass

		if flagDetach {
			pid, proxyPort, err := spawnAuthHold(port, authUser, authPass)
			if err != nil {
				teardown(absDir, st, nil)
				return err
			}
			st.AuthProxyPID = pid
			tunnelPort = proxyPort
		} else {
			localProxy, err = authproxy.Start(port, authUser, authPass)
			if err != nil {
				teardown(absDir, st, nil)
				return err
			}
			tunnelPort = localProxy.Port
		}
	}

	info("Opening Cloudflare tunnel...")
	bin, err := tunnel.Ensure()
	if err != nil {
		teardown(absDir, st, nil)
		if localProxy != nil {
			localProxy.Stop()
		}
		return err
	}
	tun, err := tunnel.Start(bin, tunnelPort, tunnelTimeout)
	if err != nil {
		teardown(absDir, st, nil)
		if localProxy != nil {
			localProxy.Stop()
		}
		return err
	}

	st.URL = tun.URL
	st.TunnelPID = tun.PID()
	st.TunnelPort = tunnelPort

	if ttl > 0 {
		exp := time.Now().Add(ttl)
		st.ExpiresAt = &exp
		if flagDetach {
			pid, err := spawnWatcher(absDir, exp)
			if err == nil {
				st.WatcherPID = pid
			}
		}
	}

	if err := state.Save(absDir, st); err != nil {
		warn("could not save state: %v", err)
	}

	ttlLabel := ""
	expiresIn := ""
	if st.ExpiresAt != nil {
		ttlLabel = ttl.String()
		expiresIn = time.Until(*st.ExpiresAt).Round(time.Minute).String()
	}

	output.PrintReady(output.ReadyOpts{
		URL:       tun.URL,
		AppPort:   port,
		Detached:  flagDetach,
		Protected: authPass != "",
		AuthUser:  authUser,
		Password:  displayPassword,
		TTL:       ttlLabel,
		ExpiresIn: expiresIn,
	})

	if flagOpen {
		_ = openURL(tun.URL)
	}

	if flagDetach {
		success("Detached. Environment keeps running in the background.")
		return nil
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	if ttl > 0 && st.ExpiresAt != nil {
		go func() {
			select {
			case <-time.After(time.Until(*st.ExpiresAt)):
				fmt.Println()
				info("TTL reached â€” tearing down...")
				teardown(absDir, st, tun)
				if localProxy != nil {
					localProxy.Stop()
				}
				success("Done.")
				os.Exit(0)
			case <-sig:
			}
		}()
	}

	<-sig
	fmt.Println()
	info("Tearing down...")
	teardown(absDir, st, tun)
	if localProxy != nil {
		localProxy.Stop()
	}
	success("Done. Environment is gone.")
	return nil
}

func resolveBuild(cfg *project.Config) bool {
	if flagNoBuild {
		return false
	}
	if cfg.Build != nil {
		return *cfg.Build
	}
	return true
}
