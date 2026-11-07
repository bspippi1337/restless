
# Reproducible Builds

Restless uses:

- -trimpath
- deterministic ldflags
- version injection via git describe

To verify:

1. Build twice in clean environments.
2. Compare sha256 checksums.
