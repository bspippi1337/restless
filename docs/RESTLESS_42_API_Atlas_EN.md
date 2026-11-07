# RESTLESS 42 API Atlas

**From one authorized domain to a source-backed API knowledge map.**

API Atlas combines a real documentation crawler, a provenance-verified library, an interactive graph, a Swagger-style method inspector, sample inputs and Postman Collection 2.1 export.

This is not a fuzzer. RESTLESS reads documentation and follows only relevant, explicitly linked documentation URLs on one exact HTTPS host. It does not brute-force endpoints or execute API operations.

## 1. Domain to library: two commands

Build the native CLI (Go 1.24 or newer):

`go build -o build/restless ./cmd/restless`

Preview a documentation source with **zero DNS or HTTP requests**:

`./build/restless library docs.your-domain.com`

Or supply an exact existing documentation URL, if you know one:

`./build/restless library https://docs.your-domain.com/reference/openapi.json --max-pages 5 --max-depth 1`

After verifying permission, scope, documentation licensing and site terms, opt in explicitly:

`./build/restless library https://docs.your-domain.com/reference/openapi.json --max-pages 12 --max-depth 2 --delay 2s --execute --authorized`

The crawler first checks robots.txt, then reads bounded and rate-limited documentation resources. It respects robots directives, avoids redirects and unrelated hosts, validates the destination IP, and stops on 429 or server errors without retrying.

Limits: 25 documentation requests, depth 3, 2 seconds minimum spacing (or longer if required by robots.txt), 2 MiB archived document size and 500 KiB robots.txt.

**Important:** GET, POST, PUT, PATCH, DELETE, HEAD, TRACE and OPTIONS found in an OpenAPI file are catalog labels, NOT methods the crawler has executed. The collector itself performs only allowed documentation GET requests.

If the site's homepage does not link to its API documentation, RESTLESS will not guess conventional paths such as /swagger.json, /api, or /v1. Provide an exact known source URL.

## 2. Generated files and provenance

Each normalized host has a separate on-disk archive:

`.restless/libraries/docs.your-domain.com/index.jsonl`: actual fetched URL, status, time, SHA-256 and content metadata.

`.restless/libraries/docs.your-domain.com/sha256/<digest>`: exact archived bytes with content-addressed filenames.

`.restless/libraries/docs.your-domain.com/catalog.json`: extracted OpenAPI operations and offline reference index.

`.restless/libraries/docs.your-domain.com/library.json`: compact API Atlas dataset with source-linked paths, methods and parameters.

The native engine validates archived document SHA-256 hashes before indexing. Subsequent conditional GETs can reuse a previously verified representation following HTTP 304.

A source hash identifies the archived documentation representation. It does not demonstrate that a server currently implements the advertised operation.

PDF resources may be archived with evidence but are not interpreted as executable API operations.

## 3. Web API Atlas

Open https://bspippi1337.github.io/restless/library/.

You can load three local JSON formats: a RESTLESS library.json, a RESTLESS catalog.json, or an OpenAPI 3 JSON specification.

The interface shows a resource graph (clickable SVG), a path/method catalog with filters, an inspector for parameters and examples, source document URLs and SHA-256, and downloads as SVG, Mermaid, Atlas JSON and Postman Collection 2.1.

Choose a method to see its documented path, declared input names/types, required parameters, auth declarations and request-body example when available. Unknown values are clearly marked as placeholders. The generated cURL is **preview text**, never executed by the page.

Postman export uses the special `{{baseUrl}}` variable, disables example query parameters by default, and includes review warnings. Importing this collection into another API client does not grant authorization.

Browser domain input only previews scope. A user may opt in to fetch one **explicit HTTPS JSON documentation URL**, but normal browser CORS rules apply and RESTLESS will not bypass them. The browser does not perform a multi-page crawl; run the native CLI/desktop and import library.json to inspect a full archive.

The web Atlas is a documentation explorer, not a live testing console or a proxy.

## 4. Native GUI on Windows, macOS and Linux

The desktop app is written using Fyne and Go as a genuine graphical desktop window, not an HTML WebView. Its source is in `desktop/`. To run it:

`cd desktop && go mod tidy && go run .`

Native graphics toolchains are needed. On Debian/Ubuntu, install gcc, libgl1-mesa-dev, xorg-dev, libwayland-dev and libxkbcommon-dev first. macOS requires Xcode Command Line Tools; Windows needs a compatible C compiler.

The desktop workflow has: domain input, scope preview, explicit consent, bounded crawl, progress log of actual archived documents, a native resource tree, method/parameter inspector, SHA-256 provenance and library import/export.

The same Fyne project targets Windows, macOS and Linux, with independent build jobs in GitHub Actions. A successful compile does not substitute for hands-on GUI tests on each platform.

## 5. Native Android companion

The Android app has a native API Atlas Activity using Android UI components and a custom Canvas-drawn resource map. You can preview a domain, copy a real CLI crawl command, and import a previously generated RESTLESS library.json through Android's system document picker to browse paths, methods and evidence.

The Android app deliberately has **no INTERNET permission**. It does not claim to perform a phone-native network crawl. Use RESTLESS desktop or CLI to create the archive, then import it.

Native Android and desktop GUIs do not run a hidden HTML page inside a WebView.

## 6. Distinguish sources from reality

**Source-backed:** operation path, method and parameter declarations extracted from verified documentation, linked to original source URLs and hashes.

**Illustrative:** synthetic input examples, placeholders, example request bodies and copied cURL snippets.

**Not verified:** live server deployment, OAuth credentials, permission to test, response timing, compatibility, and side effects.

A full API map is necessarily limited by what the documentation discloses. Private and unlinked endpoints are out of scope. This is intentional and is what distinguishes RESTLESS from a brute force discovery utility.

## 7. Testing and future work

Test the Go crawler with `go test ./internal/library ./internal/botanical`. Test the web Atlas with `node --test tests/library.test.mjs`. Native desktop platform builds are configured in `.github/workflows/desktop-atlas.yml`, and the native Android APK is built in `.github/workflows/android-native.yml`.

The current API Atlas is not a complete Postman replacement yet. There is no built-in OAuth flow, server-side collaboration, arbitrary YAML import in the browser, remote schema reference dereferencing, live execution of imported mutation methods, or automated fuzzing. It offers the durable groundwork: authorized source collection, method inventory, graph navigation, evidence and interoperable exported requests.

Sources: [RESTLESS repository](https://github.com/bspippi1337/restless), [Botanical architecture](botanical.md), [the Field Guide](RESTLESS_42_Field_Guide_EN.md).
