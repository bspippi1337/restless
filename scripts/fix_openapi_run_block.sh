#!/usr/bin/env bash
set -euo pipefail

FILE="cmd/restless-v2/openapi_cli.go"

echo "==> Rewriting clean run block"

# 1. Ensure profile import exists
if ! grep -q 'internal/profile' "$FILE"; then
  sed -i '/internal\/modules\/session/a\
\t"github.com/bspippi1337/restless/internal/profile"' "$FILE"
fi

# 2. Replace entire run case safely
awk '
BEGIN {skip=0}
/case "run":/ {
    print "case \"run\":"
    print ""
    print "    ra, sessSets, err := parseOpenAPIRunArgs(args[1:])"
    print "    if err != nil {"
    print "        fmt.Println(\"run error:\", err)"
    print "        printOpenAPIRunUsage()"
    print "        os.Exit(1)"
    print "    }"
    print ""
    print "    sess := session.New()"
    print "    for k, v := range sessSets {"
    print "        sess.Set(k, v)"
    print "    }"
    print ""
    print "    mods := []app.Module{"
    print "        sess,"
    print "        openapi.New(),"
    print "        export.New(),"
    print "    }"
    print ""
    print "    a, err := app.New(mods)"
    print "    if err != nil {"
    print "        fmt.Println(\"error:\", err)"
    print "        os.Exit(1)"
    print "    }"
    print ""
    print "    idx, err := openapi.LoadIndex(ra.ID)"
    print "    if err != nil {"
    print "        fmt.Println(\"index error:\", err)"
    print "        os.Exit(1)"
    print "    }"
    print ""
    print "    spec, err := openapi.LoadSpecFromFile(idx.RawPath)"
    print "    if err != nil {"
    print "        fmt.Println(\"spec error:\", err)"
    print "        os.Exit(1)"
    print "    }"
    print ""
    print "    // Inject profile base if not provided"
    print "    if ra.BaseOverride == \"\" {"
    print "        cfg, _ := profile.Load()"
    print "        if cfg.Active != \"\" {"
    print "            if p, ok := cfg.Profiles[cfg.Active]; ok {"
    print "                ra.BaseOverride = p.Base"
    print "            }"
    print "        }"
    print "    }"
    print ""
    print "    req, curl, err := openapi.BuildRequest(idx, spec, ra)"
    print "    if err != nil {"
    print "        fmt.Println(\"build error:\", err)"
    print "        os.Exit(1)"
    print "    }"
    print ""
    print "    if ra.ShowCurl && curl != \"\" {"
    print "        fmt.Println(curl)"
    print "    }"
    print ""
    print "    resp, err := a.RunOnce(context.Background(), req)"
    print "    if err != nil {"
    print "        fmt.Println(\"request error:\", err)"
    print "        os.Exit(1)"
    print "    }"
    print ""
    print "    fmt.Printf(\"status: %d (dur=%dms)\\n\", resp.StatusCode, resp.DurationMs)"
    print "    fmt.Println(string(resp.Body))"
    print ""
    print "    if ra.SaveAsName != \"\" {"
    print "        p, err := export.SaveJSONArtifact(ra.SaveAsName, resp)"
    print "        if err != nil {"
    print "            fmt.Println(\"save error:\", err)"
    print "            os.Exit(1)"
    print "        }"
    print "        fmt.Println(\"saved:\", p)"
    print "    }"
    skip=1
    next
}
skip && /^\s*default:/ { skip=0 }
!skip { print }
' "$FILE" > "$FILE.tmp"

mv "$FILE.tmp" "$FILE"

echo "==> Formatting"
gofmt -w "$FILE"

echo "==> Building"
go build -o restless-v2 ./cmd/restless-v2

echo "✅ openapi run block repaired cleanly."
