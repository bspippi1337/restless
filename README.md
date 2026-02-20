## Install

### Option A: Prebuilt binaries (recommended)
Restless is built by GitHub Actions. Grab the latest build artifacts here:

- **Latest workflow run artifacts:**  
  https://github.com/bspippi1337/restless/actions

Or, if you prefer tagged releases:

- **Latest release page:**  
  https://github.com/bspippi1337/restless/releases/latest

- **Current tagged release (v420):**  
  https://github.com/bspippi1337/restless/releases/tag/v420

> Note: Direct `/releases/latest/download/...` links only work when the release assets exist with matching filenames.
> Use the release page or Actions artifacts to avoid dead links.

### Option B: Build from source
```bash
git clone https://github.com/bspippi1337/restless.git
cd restless
go mod tidy
make build
./bin/restless --help
