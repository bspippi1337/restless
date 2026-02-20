#!/usr/bin/env sh
set -eu
keep="ci.yml pages.yml"
dir=".github/workflows"
[ -d "$dir" ] || exit 0

for f in "$dir"/*.yml "$dir"/*.yaml 2>/dev/null; do
  [ -e "$f" ] || continue
  base=$(basename "$f")
  case " $keep " in
    *" $base "*) : ;;
    *) echo "Removing workflow: $f"; rm -f "$f" ;;
  esac
done
echo "OK: kept $keep"
