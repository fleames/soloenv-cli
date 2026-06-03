# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Stable named subdomains (Cloudflare named tunnels)
- Remote driver for laptop-off previews
- Ephemeral database sidecar for compose

## [0.2.0] - 2026-06-03

### Added
- `--detach` to keep the environment running in the background after the command exits.
- `--protect` (auto-generated password) and `--password` for HTTP basic auth in front of the public URL, served by a local reverse proxy.
- `--ttl` for automatic teardown after a duration, with a background watcher in detached mode.
- Auto-detection and `--env-file` support for staging env files (`.env.staging`, `.env.soloenv`).
- `soloenv open` to open the staging URL in the browser, and `--open` flag on `up`.
- `soloenv logs [-f]` to stream application logs.
- QR code and clipboard copy of the URL on `up`.
- Global `--dir` flag so `status`, `logs`, `open`, and `down` can target any project folder.
- Configuration keys in `soloenv.yml`: `env_file`, `password`, `auth_user`, `ttl`.

### Changed
- Richer, boxed success output with credentials and next-step hints.
- `status` now reports auth, detached state, and TTL/expiry.

## [0.1.0] - 2026-06-03

### Added
- Initial release: `up`, `down`, `status`.
- Compose and single-Dockerfile support with automatic port detection.
- Cloudflare quick tunnel integration with automatic `cloudflared` download.
- Optional `soloenv.yml` configuration.
- Per-project `.soloenv/state.json` runtime state.

[Unreleased]: https://github.com/fleames/soloenv-cli/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/fleames/soloenv-cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/fleames/soloenv-cli/releases/tag/v0.1.0
