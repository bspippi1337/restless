# RESTLESS 42

## The Field Guide to Quiet API Discovery

**A practical, documentation-first handbook for developers, maintainers, and curious engineers**  
Edition: November 2026 | Language: English | Platform: Go CLI, Linux / Debian on WSL, macOS, and Windows builds  
Canonical repository: <https://github.com/bspippi1337/restless>  
Primary source: the RESTLESS 42 Botanical implementation in `internal/botanical/` and its CLI in `internal/cli/`.

> **The central idea:** Find out what an API already knows about itself before trying to discover it by trial and error. RESTLESS should be a careful gardener tending documented paths, not an excavator ripping up unknown systems.

**Version note:** “RESTLESS 42” is the initiative and architecture name. The version currently hard-coded in the Go CLI's `internal/version/version.go` may still display `0.4.0`; do not confuse the project's name with an automatically updated binary version. This guide describes the Botanical workflow present on `main` after the November 2026 integration and incremental-harvest work. Other older RESTLESS commands exist, but do not necessarily share all Botanical safety restrictions.

---

## Contents

1. The problem RESTLESS is trying to solve
2. What exists now, and what is still a goal
3. The five-minute mental model
4. Installing and building RESTLESS
5. The first safe session: preview, harvest, index
6. Search: API operations versus reference documents
7. Create a documentation-based probe plan
8. Compare versions with `diff`
9. Incremental harvesting: ETag, Last-Modified, and HTTP 304
10. Working with archives and provenance
11. Operating rules and responsible crawling
12. Repository Garden: cleanup and cross-platform maintenance
13. Troubleshooting and verification
14. Cookbook: three realistic developer scenarios
15. Roadmap and technical boundaries
16. Glossary and source notes
17. NotebookLM: turn this field guide into a podcast
18. Podcast narration brief and suggested questions

---

## 1. The problem RESTLESS is trying to solve

A developer returns to an API after six months and asks: “Was there an endpoint for finding orders by customer? What did I call it? Which parameter was required? Did I change the authentication?” The answer may be scattered across OpenAPI specifications, developer portals, release notes, and half-remembered terminal history.

The conventional response is often to search manually, inspect network traffic, or try plausible paths. RESTLESS 42 proposes a quieter alternative: **collect authorized public documentation, preserve its source and digest, index the explicit contracts, and suggest possible operations without executing them**.

Its guiding sequence is simple: **learn from the documentation first; plan second; verify deliberately, only with permission**. This reduces pointless guessing and makes the developer's knowledge searchable and auditable.

A second goal is historical memory. When the documentation changes, a team should be able to compare recorded API contracts and understand which documented operations were added, removed, or modified. RESTLESS is therefore an API notebook and contract cartographer as much as a probing utility.

## 2. What exists now, and what is still a goal

### Implemented Botanical features

| Command | Current purpose | Network behavior |
| --- | --- | --- |
| `restless harvest` | Preview or collect approved documentation into a local, content-addressed archive | **No requests by default**; network access requires `--execute --authorized` |
| `restless index` | Verify cached SHA-256 digests; extract explicit OpenAPI 3 operations; write `catalog.json` | Offline |
| `restless refs` | Search reference documents and whitepaper metadata | Offline |
| `restless search` | Rank documented API operations against keywords | Offline |
| `restless plan` | Build a ranked list of documented API calls for human review | Offline, **zero API operation requests** |
| `restless diff` | Compare two previously indexed snapshots of an API contract | Offline |

### Repository maintenance utilities

| Utility | Role | Important distinction |
| --- | --- | --- |
| `make garden` | Audit repository root and preview generated-directory cleanup | Dry-run; no deletions |
| `make garden-check` | Run the cross-platform Go doctor (formatting, vet, tests, build) | Does not rewrite source or run shell scripts |
| `make garden-apply` | Run the C11 cleanup tool in explicit apply mode | **Removes selected generated directories**; review first |

### Not yet delivered as an integrated Botanical capability

Embedding-based semantic search, PDF text extraction, full OpenAPI `$ref` and schema resolution, autonomous discovery across unapproved domains, automatic live probe execution, and conformance testing against running APIs are **not** implemented in this documentation-first workflow. RESTLESS's older `scan`, `discover`, `learn`, `call`, `watch`, and `fuzz` commands are separate and may initiate network requests or execute commands; do not interpret their names as a guarantee of the Botanical guardrails.

## 3. The five-minute mental model

Picture a public API documentation site as a greenhouse:

1. **Harvest.** Bring back only the permitted plant labels (documentation pages and specifications), not unknown live operations.
2. **Archive.** Store exact bytes under a SHA-256 digest. A later reader can detect whether they have changed.
3. **Index.** Read the locally stored OpenAPI 3 specification and make a catalog of documented methods, paths, summaries, parameters, and stated security requirements.
4. **Search.** Translate the developer's words into ranked possibilities: “find pet”, “list orders”, or “customer token”. This is deterministic lexical ranking, **not an LLM pretending to know a hidden endpoint**.
5. **Plan.** Produce a structured, source-linked review list. The plan is a document, not an HTTP client. It never executes the suggested operations.
6. **Compare.** Revisit a snapshot later and ask what the published API contract says has changed.

A RESTLESS catalog is therefore a **record of published claims**, not proof that each endpoint is live or accessible to a particular user. This distinction is essential.

## 4. Installing and building RESTLESS

### Requirements

- Git.
- Go **1.24 or newer**. Verify with `go version`; follow the installation instructions at <https://go.dev/dl/> if your distribution packages are older.
- Internet access only for retrieving the repository, Go modules, and any specifically authorized documentation.
- An ordinary user account is enough to run the CLI. `sudo` is needed only to install system packages, not for crawling.
- Optional: a C11 compiler and `make` for the repository-cleanup workflow.

### Debian under Windows Subsystem for Linux (WSL)

Open a Debian terminal, not Windows PowerShell, then run:

```bash
sudo apt-get update
sudo apt-get install -y git ca-certificates curl make build-essential
# Install Go 1.24+ separately if `go version` is older.
go version

git clone https://github.com/bspippi1337/restless.git "$HOME/restless-42"
cd "$HOME/restless-42"
go build -o build/restless ./cmd/restless
./build/restless --help
./build/restless version
```

If the repository already exists:

```bash
cd "$HOME/restless-42"
git switch main
git pull --ff-only
go build -o build/restless ./cmd/restless
```

`go build` may fetch declared Go dependencies. It does not contact any target API. The build output is `build/restless`. Use that exact executable in this guide so there is no ambiguity about `PATH` or stale binaries.

### Other operating systems

Go supports cross-platform builds of the main RESTLESS CLI, but the examples here use POSIX shell syntax. On macOS, use the same `git clone` and `go build` steps. On native Windows, build with `go build -o build/restless.exe ./cmd/restless` and use appropriate PowerShell equivalents for environment variables. The CLI's repository doctor is written in Go and its directory cleaner in portable C11; Debian packaging and GitHub Actions still require their own platform tooling.

**Important:** If you cloned an older integration branch in a previous session, switch to `main` as shown above. `main` now contains the merged RESTLESS and BLCKSWAN history.

## 5. Your first safe session: preview, harvest, index

This walkthrough uses a publicly known **Swagger Petstore OpenAPI example** to demonstrate the process. Public visibility alone does not grant unlimited collection rights. Read the target's terms and crawler instructions, use a deliberately small request budget, and select a different documentation source if access is disallowed.

### Step A: choose a local archive directory

```bash
cd "$HOME/restless-42"
DIR="$HOME/.local/share/restless-42/botanical"
mkdir -p "$DIR"
```

### Step B: preview the request without any network traffic

```bash
./build/restless harvest \
  https://petstore3.swagger.io/api/v3/openapi.json \
  --host petstore3.swagger.io \
  --max-pages 1 --max-depth 0 \
  --out "$DIR"
```

Without `--execute`, this is **a dry-run**: no requests, no documents, no index. The output describes scope, pacing, destination, and the flags needed for execution.

### Step C: optionally fetch one approved public specification

After independently checking access permission and site policy:

```bash
./build/restless harvest \
  https://petstore3.swagger.io/api/v3/openapi.json \
  --host petstore3.swagger.io \
  --max-pages 1 --max-depth 0 --delay 3s \
  --out "$DIR" \
  --execute --authorized
```

`--host` is an exact hostname allowlist, and the crawler stays within the **same scheme, hostname, and port**. A read of `/robots.txt` comes first. The download may be declined by site policy, a redirect, an unsupported content type, or another limit. A refusal is an expected safety result, not a reason to circumvent the restriction.

Only the documentation resource is fetched; RESTLESS does **not** automatically call the API operations described inside it.

### Step D: index the local material

```bash
./build/restless index --dir "$DIR"
```

RESTLESS reads `index.jsonl`, checks the archived content digests, parses recognized OpenAPI 3 documents, and writes `catalog.json`. Indexing is entirely local. If you skipped Step C or the harvest was blocked, `index` will report that its manifest is missing; it cannot invent a catalog.

### What to expect

The output summarizes how many documents and explicit OpenAPI operations were indexed. The exact number depends on the source and its current published contents. Do not treat counts in a tutorial as a live availability promise.

## 6. Search: API operations versus reference documents

RESTLESS keeps two different questions separate.

### Find an operation you vaguely remember

```bash
./build/restless search "find pet by id" --dir "$DIR" --limit 5
```

`search` returns JSON containing documented API operation metadata, a relevance score, and matching fields. It ranks such evidence as an operation ID, a path, a summary, a tag, and a parameter name. A high score means **a good match to the text**, not that a server was probed or that the operation succeeded.

### Find a document, paper, or guide

```bash
./build/restless refs "petstore" --dir "$DIR" --limit 10
```

`refs` searches indexed documentation records and returns source addresses, content hashes, document types, and matched terms. A PDF may appear as a preserved reference but is currently marked `pdf-unparsed`; its full body is not text-indexed by Botanical yet.

### How to read the source fields

- **`source_url` or `url`:** the exact documentation address from which a statement came.
- **`source_sha256` or `sha256`:** a fingerprint of the archived bytes, useful for integrity checks and comparing revisions.
- **`method` and `path`:** what the OpenAPI contract explicitly documents, such as `GET /pets/{petId}`.
- **`operation_id`:** the contract's symbolic operation name, if present.
- **`score`:** search relevance, not availability or permission.

If there is no match, RESTLESS returns no candidate. That is a feature: **the software should prefer an honest gap in the knowledge base to a fabricated endpoint**.

## 7. Create a documentation-based probe plan

```bash
./build/restless plan "find pet by id" --dir "$DIR" --limit 3
```

The resulting JSON has a `mode` of `DOCUMENTATION_ONLY_NO_EXECUTION` and `network_requests` equal to `0`. Each candidate identifies a method, documented path, purpose, source URL, source SHA-256, and any applicable manual-review conditions.

RESTLESS highlights situations needing additional care: non-read HTTP methods, authentication requirements stated in the contract, required parameters, path template values, and unresolved references.

For example, the published template `GET /pets/{petId}` is not a ready-to-run command: `{petId}` must be replaced with an appropriate, authorized value, and the developer must establish the target server and credentials separately. A `POST` or `DELETE` suggestion should not be executed just because it appears near the top of a ranking.

**A plan is a review artifact, not permission to issue requests.** Even `GET` endpoints can have effects in poorly designed systems. Always validate scope, data sensitivity, and operational consequences before testing a running API.

## 8. Compare versions with `diff`

RESTLESS can show what changed in the **published contract**, without contacting the server during comparison. Keep two separately indexed archive snapshots:

```bash
./build/restless index --dir "$HOME/restless-archives/before"
./build/restless index --dir "$HOME/restless-archives/after"

./build/restless diff \
  --before "$HOME/restless-archives/before" \
  --after "$HOME/restless-archives/after"
```

The input directories must already contain archived documents and their manifests; these are not created by `diff`. The output categorizes documented operations as **added**, **removed**, **changed**, or **unchanged**. Operation identity includes the source URL, HTTP method, and documented path. Differences in the source file's hash alone do not count as changes to the operation fields. A renamed source URL can therefore appear as a removal plus an addition.

A breaking-change analysis requires additional semantics and knowledge of client dependencies. `diff` reports contract changes, **not** a proven deployment state or automatic backward-compatibility verdict.

## 9. Incremental harvesting: ETag, Last-Modified, and HTTP 304

A respectful crawler should avoid downloading the same unchanged document over and over. The latest harvester can use HTTP conditional requests when the earlier archive contains a usable `ETag` or `Last-Modified` validator.

The sequence is:

1. Read previous archived metadata from the local manifest.
2. Verify that the cached file exists and matches its recorded SHA-256 before trusting it.
3. Recheck the target's `robots.txt` and continue respecting scope, rate limits, and authorization.
4. Send a conditional GET using applicable `If-None-Match` or `If-Modified-Since` headers.
5. If the site answers **304 Not Modified**, reuse the verified archived bytes rather than downloading the full body again; append a new audit entry.
6. If the document changed and the site responds **200 OK**, save the new bytes under their new digest.

A 304 without a verified cached copy is rejected. A corrupted cache is rejected. A conditional request still counts as a network request and is subject to the same rate and page budgets. This is **less wasteful crawling**, not invisible crawling.

Repeated harvests extend the manifest with additional records; they do not silently discard the original archive history. If you need independently frozen catalogs for `diff`, keep snapshots in separate directories before updating them.

## 10. Working with archives and provenance

A typical Botanical archive looks like this:

```text
botanical/
  index.jsonl            # one audit entry per archived/revalidated document
  catalog.json           # extracted, searchable documentation facts
  sha256/
    <digest>             # exact archived bytes of a document
    <another-digest>
```

`index.jsonl` is an append-oriented manifest. `catalog.json` is rebuilt from the local contents. The archive is **content-addressed**, meaning identical document bytes share the same SHA-256 filename. This is useful for reproducibility, deduplication, and detecting accidental or malicious modifications.

Provenance matters because the software must distinguish these statements:

- **Documented:** “The OpenAPI source at this URL claims `GET /pets/{petId}` exists.”
- **Derived:** “This operation is a plausible match for the search 'find pet'.”
- **Live-verified:** “An authorized request to a specific server confirmed it responds as expected.”

Botanical currently provides the first two. The third should never be silently inferred from them.

Treat cached files as data, not executable instructions. Avoid importing private credentials, customer data, or documentation you lack rights to retain.

## 11. Operating rules and responsible crawling

**RESTLESS is not a substitute for authorization or legal review.** Publicly readable pages, a permissive `robots.txt`, and a valid HTTPS response do not individually grant broad rights to reuse, republish, or test a service.

The Botanical harvester currently applies these conservative limits:

- Network activity is disabled unless **both** `--execute` and `--authorized` are supplied.
- A single permitted hostname must match the seed; discovered links stay on the same origin. Redirects are not followed.
- Production HTTP harvesting requires HTTPS, rejects credentials and query strings in seed/discovered URLs, and blocks known private, local, or reserved destination addresses.
- `robots.txt` is retrieved first; directives are respected. A 404/410 at the robots URL is handled as no robots rules under RFC 9309. Other failures block collection. Parsing is conservative and does not claim full RFC 9309 percent-encoding support.
- Default document-request budget: **5**; maximum **25**. Default link depth: **1**; maximum **3**. Minimum pacing: **2 seconds** between page requests.
- Individual archived documents are limited to **2 MiB**. The crawler stops on HTTP 429 or 5xx without retrying, and it does not follow unrelated links.
- No login automation, parameter guessing, fuzzing, active endpoint execution, or cross-domain exploration is part of `harvest`.

These are software limits, not guarantees of complete security or a binding legal interpretation. Operators must check ownership, access rules, relevant privacy obligations, licenses, and site-specific policy.

## 12. Repository Garden: cleanup and cross-platform maintenance

The RESTLESS project is deliberately keeping its live repository tidy. Fifty-three historical documents and one-off patch scripts were moved into `archive/historical/` while preserving their original Git history. This avoids confusing old migration snippets and code-rewriting shell scripts with supported maintenance tools.

The modern maintenance path has two components:

**The Go doctor** (`cmd/restless-doctor`) checks module integrity, formatting, static analysis, Go tests, and a build. It uses Go's process APIs, not an embedded Bash session. `make garden-check` runs it.

**The C11 cleaner** (`tools/c/restless-clean.c`) considers only generated directories such as `build/`, `dist/`, `bin/`, and `logs/`. It previews by default, checks the repo root, and refuses unsafe symbolic links and special files. The tool is intended for native operation across Linux, macOS, and Windows; cross-platform CI includes C compilation and safety tests.

```bash
cd "$HOME/restless-42"
make garden           # audit + cleanup preview, no deletion
make garden-check     # formatting + vet + tests + temporary build
make garden-apply     # removes selected generated directories: review first
```

The weekly Repository Garden CI job audits; it does not automatically delete files or push maintenance commits. The old scripts remain available in the historical archive for investigation, not as suggested default commands.

## 13. Troubleshooting and verification

| Symptom | Likely explanation | What to do |
| --- | --- | --- |
| `go version` reports an older release | Toolchain too old for the module | Install Go 1.24+ and confirm your `PATH` |
| `index.jsonl` not found | You only ran a dry-run, or the harvest was declined | Review the preview and any source authorization before fetching |
| Harvester refuses `robots.txt` | Target rules or retrieval errors prevent collection | Stop; choose an approved source instead of bypassing |
| `--host` rejected | Allowed hostname does not equal the seed hostname | Specify the exact source hostname and avoid redirects |
| No results from `search` | No matching OpenAPI 3 operation in the indexed source | Try alternate keywords, inspect `refs`, or add permitted official documentation |
| PDF appears but has no operation entries | PDF text parsing is not implemented | Import an official OpenAPI JSON/YAML source for operation extraction |
| Unresolved `$ref` warnings | Unsupported external, cyclic, missing, or deeper specification references | Read the specification manually; do not let the planner invent parameter values |
| 304 rejected | No verified cache for a conditional response | Inspect the local archive; reacquire only if site rules permit |
| `diff` reports everything added/removed | Different source URLs or snapshots | Confirm stable source identity across both catalogs |
| `make garden-apply` removes `build/restless` | Build output is intentionally disposable | Rebuild with `go build -o build/restless ./cmd/restless` |

Useful read-only checks:

```bash
git status --short
go version
./build/restless --help
./build/restless harvest --help
./build/restless plan --help
go test ./internal/botanical ./internal/cli
go run ./cmd/restless-doctor --audit-only
```

## 14. Cookbook: three realistic developer scenarios

### Scenario A: “I forgot the call I wrote last spring.”

Your team owns a service and publishes an OpenAPI specification. Collect the approved spec with `harvest`, build a catalog using `index`, run `search "list customer orders"`, and examine `plan "list customer orders"`. The result is a list of *documented* candidates with method, path, and parameter requirements. Check the application code or deployment separately before claiming the method works in production.

### Scenario B: “Did we break anything between the last two releases?”

Preserve a separately indexed archive for each publication, then run `diff --before ... --after ...`. Review added, removed, and modified paths with the team. Do not claim that every recorded change is backward-incompatible; a human still needs to interpret compatibility and client usage.

### Scenario C: “I need the whitepaper, not a live probe.”

Fetch or add a permitted reference document, rebuild the index, then run `refs "authentication"`. If the only source is a PDF, the archive maintains its source and digest, but the current engine will not parse its complete PDF body. Use a text/HTML version or read the PDF yourself while waiting for that feature.

## 15. Roadmap and technical boundaries

The next sensible engineering milestones are **broader OpenAPI schema resolution**, **PDF extraction with dependable citation and licensing metadata**, **improved semantic retrieval**, **incremental catalog updating**, and **opt-in, auditable verification of selected read-only operations against explicitly authorized servers**.

These are architectural directions, not delivered commands. RESTLESS should improve its memory before increasing its appetite for network traffic. High-quality context, conservative behavior, accurate provenance, and reproducible results matter more than accumulating a large request count.

A note on “intelligence”: the implemented search algorithm is lexical, with weighted text fields. It is useful, predictable, and explainable, but it does not currently understand API meaning like a large language model or vector-search system. That distinction should survive every product demo and every podcast episode.

## 16. Glossary and source notes

**API:** Application Programming Interface; a published way for software to request or exchange functionality.

**Endpoint:** A route or address that a service exposes; a path appearing in a document is not proof of a running service.

**OpenAPI:** A machine-readable contract describing HTTP operations, parameters, security expectations, and often response formats.

**HTTP GET:** A request method commonly used to retrieve information. It is not, by itself, a guarantee of zero side effects.

**`robots.txt`:** A published set of crawler-access directives, not a legal license or proof of permission.

**Fuzzing:** Sending many varied or malformed inputs to find unexpected behavior; RESTLESS 42 Botanical intentionally avoids this as a discovery strategy.

**SHA-256:** A cryptographic digest used to identify exact document bytes and check archive integrity.

**ETag / Last-Modified / 304:** HTTP mechanisms that can avoid retransmitting the body of an unchanged representation.

**Provenance:** Evidence of where information came from, often including a source URL and content hash.

**Probe plan:** A proposed set of documented operations for review. In RESTLESS Botanical, it does not send requests.

**Reference / `$ref`:** A link inside a specification to another defined schema or component; only bounded local parameter references are resolved today.

**Main implementation and reference material:**

- RESTLESS repository: <https://github.com/bspippi1337/restless>
- Botanical workflow: <https://github.com/bspippi1337/restless/blob/main/docs/botanical.md>
- Repository Garden: <https://github.com/bspippi1337/restless/blob/main/docs/repo-garden.md>
- Versioned Go CLI: <https://github.com/bspippi1337/restless/blob/main/internal/cli/root.go>
- RFC 9309 Robots Exclusion Protocol: <https://www.rfc-editor.org/rfc/rfc9309>
- OpenAPI Initiative: <https://www.openapis.org/>
- Official Go distribution: <https://go.dev/dl/>

---

## 17. NotebookLM: turn this field guide into a podcast

This handbook has been designed as a **standalone source document**. It introduces the problem, explains the actual implementation, gives concrete commands, states limits explicitly, and includes a narrative through-line. That makes it suitable for generating an English audio discussion without forcing an AI host to infer essential background.

### Quick setup

1. Open <https://notebooklm.google.com/> and create a notebook called **RESTLESS 42: Quiet API Discovery**.
2. Add this document as a **PDF** or **Markdown (`.md`)** source. Either is supported. For strongest technical grounding, you may also add the project's current `docs/botanical.md` as a separate source, but do not mix in unrelated or outdated archived docs.
3. In the Studio panel, choose **Audio Overview**, **Deep Dive**, and **English**. Select a length appropriate for a technical conversation.
4. Paste the podcast brief in Section 18 (or the separate supplied prompt file) into the customization field. Generate the discussion.
5. Listen once for technical mistakes. An AI narration is not a source of truth: cross-check claims against the source, especially permission, live verification, and unfinished features.

NotebookLM accepts Markdown and PDF sources and provides configurable Audio Overviews. Its interface and quotas may change over time; follow Google's latest Help Center instructions if labels differ.

### Suggested episode title

**“RESTLESS 42: The API Gardener Who Refused to Fuzz”**

### The narrative hook

“It's 2 a.m. A developer is staring at an API they built six months ago. Somewhere in a forest of routes is the one call they need. Do they spray requests into the darkness? Or can the documentation tell them what they forgot?”

The hosts can use that dilemma to teach the six-part pipeline: harvest, archive, index, search, plan, and compare. Let the audience discover how provenance, restraint, and replayability are engineering advantages rather than obstacles.

## 18. Podcast narration brief and suggested questions

**Format:** English-language, two-host technical Deep Dive. Target audience: API developers, DevOps engineers, security-minded maintainers, and technically curious listeners. Tone: lively, clever, and technically precise; not a fictional product launch. Prefer clear explanations to reading CLI flags aloud for minutes at a time.

**Creative direction:** Open with a developer who cannot remember their own API's endpoint names. Explain how RESTLESS 42 replaces brute-force guessing with a source-backed local index. Use the recurring **gardener versus excavator** metaphor lightly. Make the central tension compelling: intelligence can mean choosing *not* to send a request. Keep an inquisitive host who challenges unsupported claims and a technically literate host who demonstrates the workflow.

**Required technical beats:** Explain that `harvest` is dry-run by default and requires `--execute --authorized` for network access; that `robots.txt` is not a legal permission grant; that `index`, `search`, `refs`, `plan`, and `diff` are offline; that a probe plan never executes calls; and that lexical ranking is not semantic AI search. Describe SHA-256 provenance, OpenAPI operations, conditional ETags and HTTP 304, and the historical Git merge and C/Go cleanup. Distinguish documented from deployed and verified.

**Suggested arc (8–12 minutes):**

- **Cold open:** The forgotten API call, one tired engineer, and the tempting path of random requests.
- **Act 1 — The gardener:** Why legitimate documentation crawling beats blind endpoint guessing.
- **Act 2 — A memory that can be audited:** The archive, SHA-256, structured OpenAPI operation index, and reference search.
- **Act 3 — Think before touching:** Documentation-backed probe plans, safety reviews, and honest handling of missing data.
- **Act 4 — Time travel for contracts:** Offline diffs and conditional requests that spare bandwidth.
- **Act 5 — A clean toolbench:** A merged project history, Repository Garden, portable C cleanup, and a pure-Go doctor.
- **Finale:** What is already real, what is not built yet, and why responsible restraint can be a competitive feature.

**Questions one host should ask:**

1. “How can a crawler be useful if it refuses to guess undocumented endpoints?”
2. “What does a SHA-256 hash actually prove, and what does it not prove?”
3. “If RESTLESS recommends `GET /pets/{petId}`, has it verified that endpoint?”
4. “What changes between HTTP 200 and 304, and why should developers care?”
5. “How do we explain the difference between lexical ranking and genuine semantic search?”
6. “Why rewrite cleanup scripts in C and Go rather than throw more Bash at the problem?”
7. “How do access rights and robots directives differ?”
8. “What would the next genuinely valuable feature be?”

**Avoid:** Inflated performance numbers; invented customer adoption; fictional benchmarks; claims of full autonomous security scanning, fully parsed PDFs, production deployment, universally legal crawling, or live-verified API behavior. Never present an offline suggestion as the result of a successful network probe.

**Suggested closing line:** “The smartest API probe isn't always the one that makes the most requests. Sometimes it's the one that remembers where the answers were written down.”

---

*End of RESTLESS 42 Field Guide. Maintainer source: `bspippi1337/restless`. This is an educational and developer-facing guide, not legal advice or authorization to test a third-party service.*