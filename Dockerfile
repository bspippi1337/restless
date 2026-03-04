# syntax=docker/dockerfile:1.7

FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/restless ./cmd/restless

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl jq python3 tini \
 && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/restless /usr/local/bin/restless

RUN mkdir -p /demo/bin /demo/spec

RUN <<'SCRIPT'
cat > /demo/spec/openapi.json <<'JSON'
{
  "openapi":"3.0.3",
  "info":{"title":"Restless Demo API","version":"1.0"},
  "servers":[{"url":"http://127.0.0.1:8158"}],
  "paths":{
    "/health":{"get":{"responses":{"200":{"description":"ok"}}}},
    "/v1/widgets":{"get":{"responses":{"200":{"description":"widgets"}}}},
    "/v1/widgets/{id}":{"get":{"responses":{"200":{"description":"widget"}}}}
  }
}
JSON
SCRIPT

RUN <<'SCRIPT'
cat > /demo/bin/mock_api.py <<'PY'
#!/usr/bin/env python3
import json
from http.server import BaseHTTPRequestHandler,HTTPServer

widgets=[{"id":"w1","name":"flux"},{"id":"w2","name":"duck"},{"id":"w3","name":"anvil"}]

class H(BaseHTTPRequestHandler):
 def j(self,o,c=200):
  d=json.dumps(o).encode()
  self.send_response(c)
  self.send_header("Content-Type","application/json")
  self.send_header("Content-Length",str(len(d)))
  self.end_headers()
  self.wfile.write(d)

 def do_GET(self):
  p=self.path.split("?")[0]
  if p=="/health": return self.j({"ok":True})
  if p=="/v1/widgets": return self.j({"items":widgets})
  if p.startswith("/v1/widgets/"):
   i=p.split("/")[-1]
   for w in widgets:
    if w["id"]==i: return self.j(w)
   return self.j({"error":"not_found"},404)
  if p=="/openapi.json":
   return self.j(json.load(open("/demo/spec/openapi.json")))
  return self.j({"error":"unknown"},404)

 def log_message(self,*a): pass

HTTPServer(("0.0.0.0",8158),H).serve_forever()
PY
chmod +x /demo/bin/mock_api.py
SCRIPT

RUN <<'SCRIPT'
cat > /demo/bin/show.sh <<'SH'
#!/usr/bin/env bash
set -e

echo
echo "██████╗ ███████╗███████╗████████╗██╗     ███████╗███████╗███████╗"
echo "██╔══██╗██╔════╝██╔════╝╚══██╔══╝██║     ██╔════╝██╔════╝██╔════╝"
echo "██████╔╝█████╗  ███████╗   ██║   ██║     █████╗  ███████╗███████╗"
echo "██╔══██╗██╔══╝  ╚════██║   ██║   ██║     ██╔══╝  ╚════██║╚════██║"
echo "██║  ██║███████╗███████║   ██║   ███████╗███████╗███████║███████║"
echo "╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝   ╚══════╝╚══════╝╚══════╝╚══════╝"

echo
echo "Restless cinematic container demo"
echo

/demo/bin/mock_api.py &
sleep 1

echo "API preview:"
curl -s http://127.0.0.1:8158/v1/widgets | jq .

echo
echo "Restless help:"
restless --help || true

echo
echo "Request demo:"
restless GET http://127.0.0.1:8158/health || true
restless GET http://127.0.0.1:8158/v1/widgets || true
SH
chmod +x /demo/bin/show.sh
SCRIPT

ENTRYPOINT ["/usr/bin/tini","--"]
CMD ["/demo/bin/show.sh"]
