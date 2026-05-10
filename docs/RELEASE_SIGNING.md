# Release Signing Policy

Restless releases should be verifiable.

## Current release integrity

The release workflow publishes:

- release tarballs
- Debian package artifacts
- SHA256SUMS.txt
- GitHub build provenance attestations

## GPG signed tags

Stable releases should use signed tags:

```sh
git tag -s vX.Y.Z -m "restless vX.Y.Z"
git push origin vX.Y.Z
```

## Artifact signatures

When a release signing key is available, release artifacts should also be signed:

```sh
gpg --armor --detach-sign dist/restless-linux-amd64.tar.gz
gpg --armor --detach-sign dist/SHA256SUMS.txt
```

The private signing key must never be committed to this repository.

## GitHub provenance

The GitHub release workflow uses OIDC-backed build provenance attestations where supported.

## Verification

Users should verify at least one of:

- signed Git tags
- GitHub provenance attestations
- SHA256SUMS.txt
- detached GPG signatures when available
