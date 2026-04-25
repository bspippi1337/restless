# Security Policy

## Network behaviour

Restless performs bounded API discovery using safe HTTP methods only:

- GET
- HEAD
- OPTIONS

Mutation methods such as POST, PUT, PATCH, and DELETE are not used during discovery.

## Host boundary protection

Restless follows links only within the original target host.
External hosts are ignored.

## Depth limits

Traversal depth is bounded to avoid excessive scanning.

## Telemetry

Restless performs no telemetry, tracking, analytics, or background communication.

All network requests target only the user-specified API.

## Reporting issues

If you discover a security issue, please open an issue with a minimal reproduction example.
