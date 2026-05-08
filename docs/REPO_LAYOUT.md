# Restless Repository Layout

The repository is organized around a small Unix-native runtime core.

## Primary directories

### cmd/
Executable entrypoints.

### internal/
Supported runtime implementation.

### docs/
Stable user and architecture documentation.

### scripts/
Portable helper scripts and developer utilities.

## Archive areas

### archive/legacy/
Historical runtime generations preserved for reference.

### archive/experiments/
Experimental prototypes and research code.

### archive/attic/
Broken, disabled, abandoned, or incomplete fragments.

Files ending in:

- `.broken`
- `.disabled`
- `.bak`
- temporary patch artifacts

should be moved into `archive/attic/` instead of remaining in the root runtime tree.

## Repository philosophy

Restless should remain:

- terminal-native
- composable
- inspectable
- portable
- dependency-light
- Debian-friendly
- GNU/Linux-first

The root module should compile cleanly on minimal systems without requiring GUI stacks.
