# RESTLESS 42 for Android (native preview)

The RESTLESS 42 Android app is a **native Java Android application**. It does not use WebView, Electron, a JavaScript runtime or an embedded browser. The app implements the same **32 local text/API documentation operations** as the browser toolbox and can import/export compatible version-1 recipe JSON files.

> Status: test build, **not yet production-certified**. The Android CI creates a debug-signed APK for sideload testing. Debug signing does not provide a stable release signing identity.

## Features

- Native Material-style dark Android interface with RESTLESS green accents.
- Text input and output, adjustable operation recipe with move/enable/remove controls.
- One-tap examples for JSON, JWT inspection and OpenAPI 3 JSON documents.
- SHA-256/SHA-512, Base64/Base64URL, UTF-8 hex, URL encoding, HTML escaping and text tools.
- JSON, CSV, URLs, query-string conversion and unverified JWT inspection.
- Offline-only OpenAPI 3 JSON operation extraction, lexical search and probe plan generation.
- Input import via Android's system document picker; export results and recipe JSON to a user-chosen location.
- Processing occurs in a Java worker thread to avoid freezing the main UI.

## Privacy and permissions

**The manifest requests no Internet permission and no storage permission.** The app does not make HTTP requests, collect analytics, register accounts, run background services or persist input automatically.

The user can explicitly launch a separate installed browser to read online documentation. Import/export uses Android's Storage Access Framework and only files selected by the user are read or written. Copy uses Android's system clipboard, which may keep history visible to other apps.

The OpenAPI planner reads user-provided JSON and does not contact or verify any endpoint. JWT decoding is **not signature verification**.

## Build

Requirements: Android SDK platform 35 and build-tools 35.0.0, Java 17 and Gradle 8.10.2 (Android Gradle Plugin 8.7.3).

```bash
cd android
gradle --no-daemon :app:assembleDebug
```

Test APK: `android/app/build/outputs/apk/debug/app-debug.apk`

Alternatively, open the repository's **Actions** tab, select **RESTLESS 42 Native Android**, trigger it if necessary, and download the `RESTLESS-42-Android-Native-Debug` artifact from a successful run. That ZIP contains the test APK.

### Install using adb (opt-in)

```bash
adb install -r android/app/build/outputs/apk/debug/app-debug.apk
```

Android may ask for permission to sideload a debug APK. Use your own device only. If the existing app is signed with an incompatible key, Android will refuse the update; avoid uninstalling an app with unsaved local data without first exporting what you need.

## Recipe portability

Exported recipes use exactly this structure (same as the browser tool):

```json
{
  "format": "restless-42-toolbox-recipe",
  "version": 1,
  "steps": [
    {"id": "base64-decode", "option": "", "enabled": true},
    {"id": "json-pretty", "option": "", "enabled": true}
  ]
}
```

Recipes store **only the transformation steps**, not input data or credentials. Unknown IDs and recipes longer than 24 steps are rejected.

## Limits

- Input/output size: 2,000,000 UTF-16 characters, import file limit 2 MB.
- Native OpenAPI engine accepts OpenAPI 3 **JSON**, not YAML; it extracts explicit methods/paths and marks unresolved external references for review.
- JWT inspection is unverified and does not authorize access to services.
- No live HTTP client or fuzzing is included.
- This is the first native Android preview: automated Gradle compilation, APK signature/manifest checks and side-load testing are separate milestones. Manual Android UI testing on a device is still required.

## Release requirements

Before a publicly recommended stable release: test on Android 10–16, add instrumentation/UI tests, perform dependency/security review, configure a persistent private release signing key (never commit it), review accessibility, validate performance with large inputs, and publish a signed release via an appropriate release channel.

Project: https://github.com/bspippi1337/restless

MIT licensed. See `SECURITY.md` in repository root for vulnerability reports.
