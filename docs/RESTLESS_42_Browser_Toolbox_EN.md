# RESTLESS 42 Browser Toolbox

**A practical, private, composable toolkit for developer data and API documentation**

The RESTLESS 42 Browser Toolbox is a standalone client-side workbench inspired by the recipe concept of CyberChef. Its processing is implemented in plain JavaScript and native browser APIs. All transformations are performed in your browser; this version has **no server-side processor, analytics SDK, remote scripts, user account requirement, or automatic input storage**.

It complements, rather than replaces, the Go CLI. The Go program can harvest authorized documentation and preserve SHA-256 provenance. The browser toolbox transforms locally supplied text and examines OpenAPI 3 **JSON** specifications. It does not crawl websites, contact APIs, or verify that any documented endpoint is live.

## 1. Getting started

1. Open the **Toolbox** section of the RESTLESS 42 site using an HTTPS browser.
2. Choose an **Example recipe** or type a term in **Find a tool**.
3. Click an operation (or drag it) into the middle **Recipe** column.
4. Paste text into **Input**, or choose **Open file** or drag a local text file onto the input field.
5. Read **Output**. Auto-bake is enabled by default. Disable it to process manually using **Bake recipe**, or press **Ctrl/Cmd + Enter**.
6. Use **Copy**, **Save .txt**, or **Use output as input** to continue. Export/import recipes as JSON files.

Processing is deterministic except for timing. You do not need to install a browser extension or run a local server once GitHub Pages is available.

## 2. The three-column workbench

**Operations:** A searchable catalog of available functions. Click to append an operation. Drag an operation onto a recipe step to insert it before that step, or drop it into empty recipe space to append.

**Recipe:** A sequence of operations applied from top to bottom. Reorder steps by dragging or the up/down arrow buttons, disable a step temporarily, change its parameter, or delete it. At most 24 steps are accepted per recipe.

**Input and output:** Independent editable input and read-only output. Processing errors identify the failed step without silently claiming success. Use **Use output as input** to build a fresh transformation sequence. Importing/exporting a recipe never includes the input text.

Files are read via the browser's local file API. The limit is **2 MB per file**, and text operations have a **2-million-character output safety limit**. Files never upload to RESTLESS.

## 3. Operation reference

| Group | Operations | Purpose |
| --- | --- | --- |
| JSON | Pretty print, minify, sort keys, extract path | Inspect and normalize structured data |
| JSON | JSON array to CSV, CSV to JSON | Convert tabular documents |
| Encoding | To/from Base64, to/from Base64URL | Unicode-safe encoding |
| Encoding | URL component encode/decode | Work safely with percent-encoded URL data |
| Encoding | UTF-8 to/from Hex | Examine encoded bytes |
| Encoding | HTML escape/unescape | Escape and decode core HTML entities |
| Security | JWT inspect (unverified) | View JWT header and payload, **not** validate signature |
| Security | SHA-256, SHA-512 | Calculate cryptographic hashes using Web Crypto |
| Text | ROT13, uppercase, lowercase, trim | Transform simple text |
| Text | Sort lines, unique lines, text statistics | Clean and summarize text locally |
| Web | Inspect URL, query string to JSON, JSON to query string | Analyze URL and query structures |
| OpenAPI | List documented operations | Extract explicit methods and paths from an OpenAPI 3 JSON spec |
| OpenAPI | Search documented operations | Rank operations by lexical similarity to keywords |
| OpenAPI | Offline probe plan | Build a human-review checklist of documented operations |

### Supported data and conversions

All steps consume and emit **text**. Base64 and hex tools explicitly interpret bytes as UTF-8; malformed sequences fail with an error rather than silently corrupting output. For raw binary files, use a dedicated binary application instead.

**JSON path** accepts dotted property keys and numeric indexes, for example `items[0].name`; it is not a complete JSONPath engine. CSV conversion handles quoted, comma-separated, multiline text and a flat JSON array of objects. Nested JSON object values are serialized into CSV cell text.

The **JWT inspector** decodes data to help debugging but **does not authenticate tokens**. Never use it to determine whether a JWT is safe, signed by the expected key, unexpired, or authorized. Pasting sensitive JWTs or access keys into any website is a security decision; use isolated trusted environments when handling credentials.

SHA-256 and SHA-512 use the browser's native Web Crypto API. They require a secure context (HTTPS or localhost), which a properly deployed GitHub Pages URL provides.

## 4. Recipe examples

### Example A: Pretty-print an incoming JSON payload

Input:

```json
{"status":"ok","count":2,"items":[{"id":7},{"id":8}]}
```

Recipe: **JSON: pretty print**. Output is the same JSON with whitespace and indentation.

To get a compact deterministic representation, instead use **JSON: sort keys** followed by **JSON: minify**.

### Example B: Inspect an encoded string

Input: `SGVsbG8sIFJFU1RMRVNTIQ==`

Recipe: **From Base64**, followed by **Uppercase**.

Output: `HELLO, RESTLESS!`

### Example C: Examine a JWT for debugging

Choose **Inspect a JWT (unverified)** from the example recipes. Output shows JSON header and payload plus `signatureVerified: false`. The example's unsigned token is intentionally not trusted.

### Example D: Search documented operations, not guessed paths

Paste a JSON-formatted OpenAPI 3.x specification, such as:

```json
{
  "openapi": "3.0.3",
  "info": {"title": "Pets API", "version": "1"},
  "paths": {
    "/pets": {
      "get": {"operationId": "listPets", "summary": "List pets"}
    },
    "/pets/{petId}": {
      "get": {
        "operationId": "findPetById",
        "summary": "Find pet by ID",
        "parameters": [
          {"in": "path", "name": "petId", "required": true}
        ]
      }
    }
  }
}
```

Choose **OpenAPI: find operation** and enter `find pet by id` in the operation's settings, or choose **OpenAPI: offline probe plan** with the same task.

The output lists only documented operations with methods, paths, parameters, relevance scores and review warnings. It **does not create URLs, resolve unknown API hosts, send HTTP requests, test authentication, or verify server behavior**.

OpenAPI 3 JSON is supported. YAML, Swagger 2, external `$ref` fetching, embedded JavaScript execution, live API validation and arbitrary schema-resolution are not supported by the browser toolbox. Some bounded local parameter references in a `#/components/parameters/` table can be read.

### Example E: Transform CSV into JSON

Input:

```csv
name,role
"Alex, Jr.",Engineer
Sam,Researcher
```

Recipe: **CSV to JSON**. The result is an array of objects.

## 5. Recipe portability

Use **Export recipe** to download a small JSON document like this:

```json
{
  "format": "restless-42-toolbox-recipe",
  "version": 1,
  "steps": [
    {"id": "base64-decode", "enabled": true, "option": ""},
    {"id": "json-pretty", "enabled": true, "option": ""}
  ]
}
```

You can use **Import recipe** to load it again. Unknown operation IDs, unsupported versions and recipes exceeding 24 operations are rejected. The recipe file does not include your source input or credentials.

## 6. Privacy and security model

**Input never leaves this application as part of its processing flow.** The toolbox is a static website. Its operations contain no network calls; there are no external JavaScript packages or analytics services. A restrictive Content Security Policy disables network connections from the toolbox itself, including incidental fetch requests.

The **documentation and podcast** pages may link to GitHub, Google NotebookLM, and external references. Following those links leaves the toolbox and invokes the privacy rules of the destination site. Hosting providers may log ordinary page visits.

Browser-generated text is held in memory until cleared or the tab is closed. The app does not write input or recipes into localStorage or a database. The browser's own clipboard, autocomplete, swap/pagefile or download behavior remains governed by the browser and operating system.

### Responsible use

The offline planner is **not a penetration-testing authorization system**. It cannot grant permission to interact with a service, establish that an endpoint exists, or guarantee that a documented GET has no side effects. Only execute live API calls with appropriate authorization, carefully reviewed parameters, credentials and operational safeguards.

## 7. Troubleshooting

**Output is blank or a step is red:** Read the status banner at the bottom. It identifies the step that failed. Invalid JSON, malformed encoding, an unsupported file type or a missing JSON path can cause errors.

**SHA digest does not work:** Use HTTPS or localhost. Web Crypto may be unavailable in insecure contexts.

**A pasted OpenAPI YAML document fails:** Convert it to JSON in a trusted tool first. The current OpenAPI browser operations accept JSON only.

**An OpenAPI path is missing:** The browser indexes only explicit HTTP operations in `paths`; it does not invent endpoints and does not resolve referenced path objects.

**An untrusted JWT displays a payload:** This is expected. Decoding is not verification.

**Auto-bake is slow:** Disable auto-bake, shorten the input or simplify the recipe. A single step or file is bounded to prevent accidental excessive processing. Large arbitrary files are not supported.

**Copy is blocked:** Your browser may disallow programmatic clipboard access. Manually select output text or use **Save .txt**.

**Nothing is saved after a reload:** This is intentional. To preserve a pipeline, export it as a recipe JSON file first.

## 8. Source, testing and limitations

The browser toolbox implementation is split into a pure transform module (`site/toolbox/transforms.mjs`) and UI module (`site/toolbox/app.mjs`). Independent automated tests live in `tests/toolbox.test.mjs`. The GitHub Pages workflow executes those tests, builds the complete English field guide, and validates local links before deployment.

A successful GitHub Actions *build* is not proof that GitHub Pages itself is enabled. If publishing fails with HTTP 404, the repository owner must set **Settings → Pages → Source: GitHub Actions**.

For the complete CLI documentation, see [The Field Guide to Quiet API Discovery](https://github.com/bspippi1337/restless/blob/main/docs/RESTLESS_42_Field_Guide_EN.md). For source, issues and license, see [bspippi1337/restless](https://github.com/bspippi1337/restless).

**RESTLESS 42 is the project initiative name; it should not be mistaken for the current binary's semantic version.**
