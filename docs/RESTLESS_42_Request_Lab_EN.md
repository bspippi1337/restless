# RESTLESS 42 Request Lab

**One explicitly authorized HTTPS request. No probing spiral.**

Request Lab is the deliberately **network-capable** complement to RESTLESS 42's offline Browser Toolbox and Contract Lab. It runs entirely as a static page in your browser. Unlike the other two apps, it can make a network request, but **only** after you supply the complete URL, separately approve the exact hostname, confirm you have permission, and press **Send approved request**.

Request Lab is useful for inspecting the behavior of an API you own or are explicitly authorized to access. It is not a scanning service, proxy, fuzzing framework or method for bypassing CORS.

## 1. Start with a dry preview

1. Open the RESTLESS 42 **Request Lab**.
2. Click **Load a non-executing example**. The example is a public API documentation URL, but no network request happens.
3. Read the target URL and its separate exact **Approved hostname**.
4. Leave permission **unchecked** while inspecting the example.
5. If and only if you have permission and have verified the target's policies, check the permission confirmation and press **Send approved request**.

The button cannot be used without the permission checkbox and a URL and host. The JavaScript engine independently validates permission, HTTPS, hostname and method on every execution. URL fragments and embedded credentials are refused.

**A site’s public availability, a robots.txt allow rule or the ability to load a URL do not constitute authorization.**

## 2. Supported requests

Request Lab permits precisely **GET, HEAD and OPTIONS**. POST, PUT, PATCH, DELETE, TRACE and other methods are rejected. No request body is sent.

Network behavior is bounded:

| Control | Behavior |
| --- | --- |
| Destination | HTTPS URL, publicly named DNS hostname |
| Scope | Exact approved hostname must match URL hostname |
| Methods | GET, HEAD, OPTIONS only |
| Credentials | Browser credentials and cookies are never sent; manually entered Authorization header requires an extra opt-in |
| Redirects | Error: redirect following disabled |
| Retry | None |
| Referrer | Suppressed |
| Browser CORS | Enforced; no proxy or workaround |
| Timeout | 1–15 seconds; defaults to 10 |
| Response size | Maximum 2 MB, including decoded body |
| Headers | Up to 20 custom string headers in an 8 KB JSON object |
| Persistence | No automatic storage, telemetry, tracking or backend processor |

A request is initiated only by a human action. Response results remain in memory until you clear or close the page.

## 3. Custom headers and credential safety

Write custom headers as a JSON object:

```json
{
  "Accept": "application/json",
  "X-Client-Id": "my-approved-test-client"
}
```

Header names and values are validated. Browser-controlled or unsafe headers, such as `Host`, `Cookie`, `Origin`, `Referer`, `Sec-*` and `Proxy-*`, are rejected. Newlines in values are not accepted.

A manually entered `Authorization` header is **only** permitted after enabling the separate credentials checkbox. That checkbox does not change browser cookie handling: `credentials: "omit"` is always used. Treat pasted credentials as sensitive, consider clipboard privacy, and do not use long-lived secrets in a browser if an isolated development environment would be safer.

The exported report contains **response evidence**, not request header values. Common sensitive URL query parameters, including access tokens, passwords, API keys and signatures, are redacted from the exported evidence URL. That is a best-effort safety feature, not a guarantee that arbitrary secret-bearing query names can be recognized. Inspect the export before sharing.

## 4. Browser CORS limitations

A remote API needs to allow this GitHub Pages origin through its CORS response headers. If it does not, the browser blocks inspection even if you own the API. RESTLESS **does not bypass CORS**, relay your request through a third-party proxy, or silently widen the hostname scope.

A CORS error, failed TLS validation, DNS failure and other browser network failures may look similar. The UI reports the uncertainty instead of claiming an endpoint was down. For an API that is not browser-accessible, use properly authorized server-side developer tools instead.

## 5. Interpreting the response

For a completed request, the UI displays:

- HTTP status, status text and elapsed time.
- Number of downloaded response bytes, bounded to 2 MB.
- SHA-256 digest computed from actual received bytes.
- The response body, pretty-formatted when it is valid JSON.
- Response headers the browser makes available to JavaScript.

Binary responses show a bounded hexadecimal preview rather than trying to present raw bytes as printable text.

A 200 response does not prove an API is safe, supported or authorized. A 404 is an observed response at a moment in time, not proof that no endpoint exists anywhere. A 429 or 500 is recorded without automatic retries. GET is conventionally used to read, but application implementations can have side effects.

## 6. Distinguishing the three browser tools

**Toolbox:** a private CyberChef-style text/encoding/JSON/OpenAPI recipe engine. Its Content Security Policy explicitly has `connect-src 'none'`; no API requests are sent.

**Contract Lab:** a private OpenAPI JSON explorer and version comparison engine. It identifies documented changes and potential compatibility review risks. It also has `connect-src 'none'`.

**Request Lab:** an explicit HTTPS request console with `connect-src https:` and consent/host controls. It makes **one** operator-approved request when the Send button is pressed.

These should never be conflated: a **document-based probe plan is not a live probe**, and a live response to one authorized URL does not validate an entire API specification.

## 7. Developer usage and testing

The local processing module is in `site/request/engine.mjs`, with browser UI in `site/request/app.mjs`. Node 22 native tests are in `tests/request.test.mjs`:

```sh
node --check site/request/engine.mjs
node --check site/request/app.mjs
node --test tests/request.test.mjs
```

Tests verify permission gates, method restrictions, public DNS hostname allowlist, custom headers, no request retries, byte caps, timeouts, SHA-256 evidence and CORS/network failure handling using mocked responses. They do **not** send live HTTP traffic to an external API during CI.

For general RESTLESS 42 architecture, see [The Field Guide to Quiet API Discovery](../../guide/). For the native Android application, see [Android native app guide](https://github.com/bspippi1337/restless/blob/main/android/README.md).

## 8. Troubleshooting

**Send is disabled:** provide a URL, approved hostname and explicitly check the permission box.

**Approved hostname mismatch:** enter only the hostname, not `https://` or a path. The URL hostname must match exactly.

**Browser request failed:** check that the service allows browser CORS for this origin and that HTTPS and the host are valid. No hidden proxy is available.

**Authentication header refused:** enable the separate credential confirmation only if your security policy permits testing with that credential.

**Large response rejected:** this is intentional. Use a different, explicitly authorized workflow for larger exports.

**Browser output changed unexpectedly:** JSON pretty-printing changes whitespace, but the SHA-256 digest is computed from the original received bytes, not the formatted display text.

**Remember:** the safest request is the one you understand and have permission to send.
