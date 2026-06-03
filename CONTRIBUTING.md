# Contributing to SoloEnv

Thanks for your interest in making SoloEnv better! This project is for solo devs, by people who like small, sharp tools. Contributions of all sizes are welcome.

## Ways to contribute

- ðŸ› **Report bugs** with the [bug report template](https://github.com/fleames/soloenv-cli/issues/new?template=bug_report.yml)
- ðŸ’¡ **Request features** with the [feature request template](https://github.com/fleames/soloenv-cli/issues/new?template=feature_request.yml)
- ðŸ“– **Improve docs** â€” typos and clarifications are great first PRs
- ðŸ§‘â€ðŸ’» **Send code** â€” see below

## Development setup

You need [Go](https://go.dev/dl/) 1.26+ and [Docker](https://docs.docker.com/get-docker/).

```bash
git clone https://github.com/fleames/soloenv-cli.git
cd soloenv-cli
make build      # builds ./soloenv (or soloenv.exe on Windows)
make test       # run the test suite
make check      # gofmt check + go vet + tests
```

To try your build against a real app:

```bash
cd /path/to/some-docker-app
/path/to/soloenv-cli/soloenv up
```

## Project layout

```
cmd/                CLI commands (cobra) and OS-specific helpers
internal/
  authproxy/        local HTTP basic-auth reverse proxy
  docker/           docker / docker compose wrappers (os/exec)
  output/           terminal output, QR code, clipboard
  preflight/        environment checks
  project/          project detection, config, env/auth/ttl resolution
  state/            .soloenv/state.json read/write
  tunnel/           cloudflared management and URL parsing
templates/          drop-in CI templates (PR previews)
```

## Pull request guidelines

1. **Open an issue first** for anything non-trivial so we can agree on direction.
2. Keep PRs focused â€” one logical change per PR.
3. Run `make check` before pushing; CI runs `gofmt`, `go vet`, and tests on Linux, macOS, and Windows.
4. Add or update tests for behavior changes.
5. Update [`CHANGELOG.md`](CHANGELOG.md) under `## [Unreleased]`.
6. Use clear commit messages (we loosely follow [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).

## Code style

- Standard `gofmt` formatting (no manual alignment fights).
- Prefer the standard library; add dependencies only when they clearly earn their place.
- Comments explain **why**, not what.
- Keep the CLI output friendly and actionable â€” error messages should suggest a fix.

## Reporting security issues

Please do **not** open public issues for security problems. See [SECURITY.md](SECURITY.md).

By contributing, you agree that your contributions are licensed under the [MIT License](LICENSE).
