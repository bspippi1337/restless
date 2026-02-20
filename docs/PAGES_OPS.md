# GH Pages Ops

## What to keep
- `.github/workflows/pages.yml` (deploys docs/)
- `.github/workflows/ci.yml` (build/test)

## What to delete
Any experimental workflows that modify Pages or push to `gh-pages` branches.
If something goes weird, run:

```sh
sh scripts/cleanup-workflows.sh
```

## How to verify
1) Repo → Settings → Pages → Source: GitHub Actions
2) Actions tab: workflow **pages** should be green
3) Site: https://bspippi1337.github.io/restless/
