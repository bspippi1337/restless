# Security & Safety

Restless v2 is designed for **developer onboarding** and **safe discovery**.

## Default safety
- Only uses GET/HEAD/OPTIONS by default
- Hard budgets: max time, max pages, max endpoints per phase
- Treats 401/403 as endpoint existence signals
- Avoids aggressive crawling

## What Restless does NOT do
- brute force endpoint enumeration
- bypass authentication
- send write calls without explicit user action

## Reporting
If you believe Restless behaves unsafely, open an issue with:
- domain tested
- command run
- logs (redact secrets)
