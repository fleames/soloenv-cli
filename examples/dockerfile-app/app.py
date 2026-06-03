"""Minimal stdlib web app that echoes an env var, to demo .env.staging injection."""
import os
from http.server import BaseHTTPRequestHandler, HTTPServer

GREETING = os.environ.get("GREETING", "Hello (no env set)")
PORT = int(os.environ.get("PORT", "8000"))

PAGE = """<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>SoloEnv Dockerfile example</title>
<style>
body{{margin:0;min-height:100vh;display:grid;place-items:center;
font-family:system-ui,-apple-system,"Segoe UI",sans-serif;background:#0b0d13;color:#eef1f8}}
.card{{text-align:center;padding:2rem 2.5rem;border:1px solid #252b3d;
border-radius:16px;background:#141826}}
.accent{{color:#7fd962}}p{{color:#9aa4ba}}
</style></head><body>
<div class="card">
<h1><span class="accent">{greeting}</span></h1>
<p>This value came from <code>.env.staging</code> via SoloEnv.</p>
</div></body></html>"""


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.end_headers()
        self.wfile.write(PAGE.format(greeting=GREETING).encode())

    def log_message(self, *args):
        pass


if __name__ == "__main__":
    print(f"listening on :{PORT} with GREETING={GREETING!r}")
    HTTPServer(("0.0.0.0", PORT), Handler).serve_forever()
