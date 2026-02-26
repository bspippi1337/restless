#!/usr/bin/env bash
set -euo pipefail

MAIN="cmd/restless/main.go"

echo "==> Making --spec optional and adding surface fallback..."

perl -0777 -i -pe '
s/if \*spec == "" \|\| \*base == "" \{\s*fmt\.Println\("missing --spec or --base"\)\s*fs\.Usage\(\)\s*os\.Exit\(2\)\s*\}/if *base == "" {\n\t\tfmt.Println("missing --base")\n\t\tfs.Usage()\n\t\tos.Exit(2)\n\t}/s
' "$MAIN"

perl -0777 -i -pe '
s/rep, err := validate\.Run\(ctx, validate\.Options\{[\s\S]*?\}\)\n\s*if err != nil \{[\s\S]*?\}\n\n\s*snap := snapshot\.FromValidateReport\(\*base, \*spec, rep\)/var snap snapshot.Snapshot\n\n\tif *spec != "" {\n\t\trep, err := validate.Run(ctx, validate.Options{\n\t\t\tSpecPath:   *spec,\n\t\t\tBaseURL:    *base,\n\t\t\tTimeout:    time.Duration(*timeout) * time.Second,\n\t\t\tStrictLive: *strict,\n\t\t})\n\t\tif err != nil {\n\t\t\tfmt.Println("validate error:", err)\n\t\t\tos.Exit(2)\n\t\t}\n\t\tsnap = snapshot.FromValidateReport(*base, *spec, rep)\n\t} else {\n\t\t// Surface mode fallback\n\t\tfmt.Println("No --spec provided. Running surface snapshot...")\n\t\treq := types.Request{\n\t\t\tMethod:  "GET",\n\t\t\tURL:     *base,\n\t\t\tHeaders: http.Header{},\n\t\t}\n\t\ta := buildApp()\n\t\tresp, err := a.RunOnce(ctx, req)\n\t\tif err != nil {\n\t\t\tfmt.Println("surface error:", err)\n\t\t\tos.Exit(2)\n\t\t}\n\t\tsnap = snapshot.Snapshot{\n\t\t\tKind:      "restless.snapshot.v1",\n\t\t\tCreatedAt: time.Now().UTC().Format(time.RFC3339),\n\t\t\tBaseURL:   *base,\n\t\t\tSpecPath:  "",\n\t\t\tChecked:   1,\n\t\t\tFailed:    0,\n\t\t\tEndpoints: []snapshot.Endpoint{{Method: "GET", Path: "/", ActualCode: resp.StatusCode}},\n\t\t}\n\t\tsnap.Fingerprint = snapshot.Fingerprint(snap)\n\t}\n/s
' "$MAIN"

echo "==> Formatting..."
gofmt -w cmd internal

echo "==> Building..."
go build ./cmd/restless

echo "==> Committing..."
git add -A
git commit -m "feat(snapshot): make --spec optional with surface fallback" || true

echo
echo "✓ Snapshot now works without --spec."
echo
echo "Try:"
echo "  restless snapshot --base https://api.github.com --out github.snap.json"
