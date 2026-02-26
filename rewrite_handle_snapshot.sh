#!/usr/bin/env bash
set -euo pipefail

FILE="cmd/restless/main.go"

[ -f "$FILE" ] || { echo "main.go not found"; exit 1; }

echo "==> Rewriting handleSnapshot() safely..."

awk '
BEGIN { skip=0 }
/^func handleSnapshot\(args \[\]string\)/ {
    skip=1
    print "func handleSnapshot(args []string) {"
    print "    fs := flag.NewFlagSet(\"snapshot\", flag.ExitOnError)"
    print ""
    print "    spec := fs.String(\"spec\", \"\", \"Path to OpenAPI spec (optional)\")"
    print "    base := fs.String(\"base\", \"\", \"Base URL (required)\")"
    print "    out := fs.String(\"out\", \"snapshot.json\", \"Output file\")"
    print "    timeout := fs.Int(\"timeout\", 7, \"Timeout in seconds\")"
    print "    strict := fs.Bool(\"strict\", false, \"Strict mode\")"
    print ""
    print "    fs.Parse(args)"
    print ""
    print "    if *base == \"\" {"
    print "        fmt.Println(\"missing --base\")"
    print "        fs.Usage()"
    print "        os.Exit(2)"
    print "    }"
    print ""
    print "    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)"
    print "    defer cancel()"
    print ""
    print "    var snap snapshot.Snapshot"
    print ""
    print "    if *spec != \"\" {"
    print "        rep, err := validate.Run(ctx, validate.Options{"
    print "            SpecPath:   *spec,"
    print "            BaseURL:    *base,"
    print "            Timeout:    time.Duration(*timeout) * time.Second,"
    print "            StrictLive: *strict,"
    print "        })"
    print "        if err != nil {"
    print "            fmt.Println(\"validate error:\", err)"
    print "            os.Exit(2)"
    print "        }"
    print "        snap = snapshot.FromValidateReport(*base, *spec, rep)"
    print "    } else {"
    print "        fmt.Println(\"No --spec provided. Running surface snapshot...\")"
    print ""
    print "        req := types.Request{"
    print "            Method:  \"GET\","
    print "            URL:     *base,"
    print "            Headers: http.Header{},"
    print "        }"
    print ""
    print "        sess := session.New()"
    print "        mods := []app.Module{"
    print "            sess,"
    print "            openapi.New(),"
    print "            export.New(),"
    print "            bench.New(),"
    print "        }"
    print ""
    print "        a, err := app.New(mods)"
    print "        if err != nil {"
    print "            fmt.Println(\"ERROR:\", err)"
    print "            os.Exit(2)"
    print "        }"
    print ""
    print "        resp, err := a.RunOnce(ctx, req)"
    print "        if err != nil {"
    print "            fmt.Println(\"surface error:\", err)"
    print "            os.Exit(2)"
    print "        }"
    print ""
    print "        snap = snapshot.Snapshot{"
    print "            Kind:      \"restless.snapshot.v1\","
    print "            CreatedAt: time.Now().UTC().Format(time.RFC3339),"
    print "            BaseURL:   *base,"
    print "            Checked:   1,"
    print "            Failed:    0,"
    print "            Endpoints: []snapshot.Endpoint{{"
    print "                Method:     \"GET\","
    print "                Path:       \"/\","
    print "                ActualCode: resp.StatusCode,"
    print "            }},"
    print "        }"
    print "        snap.Fingerprint = snapshot.Fingerprint(snap)"
    print "    }"
    print ""
    print "    if err := snapshot.WriteJSON(*out, snap); err != nil {"
    print "        fmt.Println(\"write error:\", err)"
    print "        os.Exit(2)"
    print "    }"
    print ""
    print "    snapshot.PrintHuman(os.Stdout, snap)"
    print "    fmt.Println(\"wrote:\", *out)"
    print "}"
    next
}

skip==1 && /^\}/ { skip=0; next }
skip==1 { next }

{ print }
' "$FILE" > "$FILE.tmp"

mv "$FILE.tmp" "$FILE"

echo "==> Formatting..."
gofmt -w cmd

echo "==> Building..."
go build ./cmd/restless

echo "==> Done."
echo
echo "Test with:"
echo "  restless snapshot --base https://api.github.com"
