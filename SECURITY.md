# Security Policy

## Supported versions

SoloEnv is pre-1.0 and moves fast. Security fixes land on the latest released
minor version.

| Version | Supported |
|---------|-----------|
| 0.2.x   | âœ… |
| < 0.2   | âŒ |

## Reporting a vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, use one of the following:

- GitHub's [private vulnerability reporting](https://github.com/fleames/soloenv-cli/security/advisories/new)
- Email **security@soloenv.dev**

Please include:

- A description of the issue and its impact
- Steps to reproduce (a minimal `compose.yaml`/`Dockerfile` helps)
- The SoloEnv version (`soloenv --version`) and your OS

We aim to acknowledge reports within 72 hours and to provide a remediation
timeline after triage.

## Security model notes

- SoloEnv exposes your local app to the public internet via a Cloudflare quick
  tunnel. Treat any running environment as public unless you use `--protect`
  or `--password`.
- Basic auth protection is transport-secured by the tunnel's HTTPS, but it is
  still basic auth â€” use strong, single-use passwords for shared previews.
- `cloudflared` is downloaded from the official Cloudflare GitHub releases on
  first use and cached in your user cache directory.
- Runtime state in `.soloenv/state.json` may contain the auth username and a
  reference to a generated password's presence; keep `.soloenv/` out of version
  control (it is gitignored by default).
