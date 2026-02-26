#!/usr/bin/env bash
set -euo pipefail

MAIN="cmd/restless/main.go"

[ -f "$MAIN" ] || { echo "main.go not found"; exit 1; }

echo "==> Patching switch..."

# Add case "validate" if not present
if ! grep -q 'case "validate"' "$MAIN"; then
  perl -0777 -i -pe '
    s/(case "profile":\s*handleProfile\(os\.Args\[2:\]\)\s*return\s*)/$1\n\t\tcase "validate":\n\t\t\thandleValidate(os.Args[2:])\n\t\t\treturn\n/s
  ' "$MAIN"
fi

echo "==> Adding validate import..."

if ! grep -q 'internal/validate' "$MAIN"; then
  perl -0777 -i -pe '
    s|(github\.com/bspippi1337/restless/internal/modules/session"\n)|$1\t"github.com/bspippi1337/restless/internal/validate"\n|
  ' "$MAIN"
fi

echo "==> Adding handleValidate()..."

if ! grep -q 'func handleValidate' "$MAIN"; then
cat >> "$MAIN" <<'GO'

func handleValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)

	spec := fs.String("spec", "", "Path to OpenAPI spec")
	base := fs.String("base", "", "Base URL")
	strict := fs.Bool("strict", false, "Strict mode")
	jsonOut := fs.Bool("json", false, "JSON output")

	fs.Parse(args)

	if *spec == "" || *base == "" {
		fmt.Println("missing --spec or --base")
		fs.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	rep, err := validate.Run(ctx, validate.Options{
		SpecPath:   *spec,
		BaseURL:    *base,
		StrictLive: *strict,
	})
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	if *jsonOut {
		validate.PrintJSON(rep, os.Stdout)
	} else {
		validate.PrintHuman(rep, os.Stdout)
	}

	if !rep.OK {
		os.Exit(1)
	}
}
GO
fi

echo "==> Formatting..."
gofmt -w "$MAIN"

echo "==> Building..."
go build ./cmd/restless

echo "==> Committing..."
git add "$MAIN"
git commit -m "feat(validate): wire validate subcommand" || true

echo
echo "✅ validate wired successfully."
echo
echo "Test with:"
echo "  restless validate --spec openapi.yaml --base https://api.example.com"
