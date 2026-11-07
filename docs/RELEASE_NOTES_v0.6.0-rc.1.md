# RESTLESS 42 v0.6.0-rc.1 | API Atlas

This is a release candidate for RESTLESS 42, a documentation-first API knowledge workstation. It combines native source crawling, cryptographically traceable catalogs, interactive maps, and examples that are never executed automatically.

## New capabilities

- **Browser API Atlas:** bespoke visual identity and vector topology, searchable OpenAPI resource graph, parameter editor, response inspector, evidence references and zoom/fit controls.
- **Examples:** locally generated cURL, JavaScript fetch, and Python requests; Postman Collection 2.1, Mermaid graph, SVG map and JSON exports.
- **Real crawler:** the new RESTLESS library CLI and native desktop use a bounded, opt-in, robots-aware, exact-host HTTPS documentation collector. It archives source bytes with SHA-256 and indexes declared HTTP operations offline. It does not guess endpoints or fuzz.
- **Native desktop:** dedicated Fyne graphical app for Windows, macOS and Linux with domain input, preview, crawl progress, source tree, documented request examples and clipboard copying.
- **Native Android:** Java and Android Canvas app with offline library import and documented resource mapping. It intentionally has no Internet permission.
- **Developer trust:** security regression tests, cross-platform builds, browser QA, SHA-256 manifest and complete English documentation.

## Important boundaries

The Android APK is debug-signed for sideloading and the desktop executables are not notarized or app-store signed. Physical device testing remains required.

The Android companion does not perform network crawling. The browser page honors ordinary CORS and can fetch only one explicit HTTPS JSON document with authorization; the complete documentation crawl runs in the CLI or desktop.

The operation examples are not live requests. Documentation is not a guarantee of deployment or permission, and full API coverage is not possible when documentation is unlinked or private.

Version v0.6.0-rc.1 means pre-release. Verify published assets against SHA256SUMS.txt.

Project: https://github.com/bspippi1337/restless
API Atlas: https://bspippi1337.github.io/restless/library/
Design and implementation PR: https://github.com/bspippi1337/restless/pull/19
