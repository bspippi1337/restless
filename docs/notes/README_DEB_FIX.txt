Restless Debian Packaging Fix
=============================

This patch fixes dpkg version errors when using git-describe.

Example conversion:

v6.0.0-16-gc68d817-dirty
→
6.0.0+git16.gc68d817

Steps:

1. Unzip into repo root
2. Ensure script is executable

   chmod +x scripts/build_deb.sh

3. Run

   make deb

Result:

dist/restless_<version>_amd64.deb
