# SoloEnv examples

Tiny, self-contained apps you can use to try SoloEnv in seconds. Each folder is a
complete project with everything `soloenv up` needs.

> Build the CLI first (from the repo root): `go build -o soloenv .`
> Then run the commands below from inside an example folder.

| Example | What it shows |
|---------|---------------|
| [`static-site/`](static-site/) | Compose project (nginx) serving static files, auto port detection, `soloenv.yml` |
| [`dockerfile-app/`](dockerfile-app/) | Single `Dockerfile` with `EXPOSE`, plus `.env.staging` injection |

## static-site (compose)

```bash
cd static-site
soloenv up
# open the printed https://*.trycloudflare.com URL
```

Try it protected and self-expiring:

```bash
soloenv up --protect --ttl 30m --open
```

## dockerfile-app (Dockerfile + env)

A minimal Python app that reads `GREETING` from the environment, demonstrating
how `.env.staging` flows into the preview.

```bash
cd dockerfile-app
soloenv up
# the page shows the GREETING value from .env.staging
```

When you're done in either example:

```bash
soloenv down
```
