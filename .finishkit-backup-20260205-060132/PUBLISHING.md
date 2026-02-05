# Publishing (one-time + release flow)

## One-time setup
1. Create an npm token that can publish the package.
2. Add it to GitHub repo secrets as **NPM_TOKEN**.
3. Ensure the npm package name is available (default: `restless-uac`).

## Release flow (what you do)
1. Bump versions:
   - `npm/package.json` version: `X.Y.Z`
2. Commit and push.
3. Tag and push:
   ```sh
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
4. GitHub Actions will:
   - build and upload release assets via GoReleaser
   - publish the npm wrapper when the GitHub Release is published

## Notes
- The npm installer downloads assets named like:
  `restless_X.Y.Z_<os>_<arch>.tar.gz`

## Homebrew tap setup (one-time)
1) Create a tap repository named:
   - `homebrew-tap` under your GitHub user/org (e.g. `blckswan1337/homebrew-tap`)
2) Add a secret in the main repo:
   - `GORELEASER_TOKEN` = GitHub classic PAT with `repo` scope (allows pushing formula updates to the tap repo)

After you push a tag `vX.Y.Z`, GoReleaser will update `Formula/restless.rb` in the tap repo automatically.
