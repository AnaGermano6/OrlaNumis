# Security Principles

> The security posture we intentionally enforce across Orla Numis. This is the stable controls catalog; the threat inventory lives in [threat-model.md](threat-model.md) and the operational playbook lives in [incident-response.md](incident-response.md).

## Scope and posture

Orla Numis is a self-hosted, three-service application with a relatively small operator footprint but a broad attack surface: browser UI, Go API, Python agent, container build chain, and CI/CD. The June 2025 review documented 24 enumerated findings across backend, frontend, and supply-chain surfaces (the legacy executive summary says 25); the highest-risk themes were browser trust boundaries, token exposure, input validation, and mutable build dependencies.

The project already starts from several strong defaults: layered architecture in the Go API, modern authentication primitives (JWT, refresh rotation, lockout), minimal production containers, lockfiles, and least-privilege CI permissions. This document captures the controls we expect every future change to preserve or strengthen.

## Principles we follow

1. **Single-ingress trust boundary** — the Vue SPA talks only to the Go API, and the Go API is the only HTTP client of the Python agent. This reduces duplicate auth, CORS, and rate-limit surfaces.
2. **Thin handlers, centralized data access** — request parsing stays in handlers and database access stays in repositories, which keeps validation and query hardening reviewable.
3. **Modern auth with revocation** — short-lived JWT access tokens, rotated refresh tokens, account lockout and automatic IP blocking are the baseline. Username and password is the only sign-in method.
4. **Minimal production artifacts** — multi-stage Docker builds, a static Go binary, and small runtime images reduce attack surface.
5. **Secrets stay out of git history** — credentials belong in environment variables or ignored local files, never in tracked config.
6. **Generated or cached artifacts are not trusted by default** — Swagger examples, build output, and AI-generated content must be treated as potentially misleading or unsafe until validated.
7. **Least privilege by default** — CI permissions, runtime users, and future service credentials should be scoped to the minimum required capability.
8. **Security findings stay actionable** — tracked threats carry a severity, location, remediation path, and status in [threat-model.md](threat-model.md); operational handling lives in [incident-response.md](incident-response.md).

## Control areas

| Control area | What we enforce |
|---|---|
| Input validation | User-controlled query inputs must be parameterized or validated against explicit allowlists before they reach SQL, search parameters, or downstream services. |
| Output encoding and rendering | Untrusted HTML or Markdown rendered in the browser must be sanitized before `v-html` or equivalent sinks. |
| Secret handling | Secrets, API keys, and signing material must come from environment or secret stores; tracked files may only contain obviously fake examples. |
| Authentication and token policy | Access tokens stay short-lived, refresh tokens rotate on use and are stored hashed, failed sign-ins lock the account and block the IP for a while, and registration stays closed after the first user. |
| Internal agent boundary | The Python agent is reachable only from the Go API, with the shared service token; its callbacks into the API use short-lived, owner- and job-scoped credentials. |
| Upload safety | File uploads validate extension and MIME/magic bytes, and body sizes stay capped to prevent memory exhaustion. |
| Abuse resistance | Auth endpoints are rate-limited and expensive operations should be reviewed for throttling or queueing. |
| Transport security | Production deployments use HTTPS for browser ↔ API traffic. Tokens are never designed around plaintext transport assumptions. |
| Container hardening | Production containers should pin trusted base images and run as non-root users. |
| Supply chain integrity | GitHub Actions pin by SHA, lockfiles stay committed, dependency updates are reviewed, and secret scanning runs before commits. |
| Incident readiness | Security reports are acknowledged through [SECURITY.md](../SECURITY.md), triaged with [incident-response.md](incident-response.md), and tracked against the live threat inventory. |

## Current implementation expectations

### Backend and API

- Prefer repository scopes and parameter binding over dynamic query fragments.
- Fail safely: log internal errors server-side and return generic client messages.
- Set multipart and JSON body caps for endpoints that accept user payloads.
- Keep the full backups out of the web root and readable only by administrators; they hold password hashes and saved provider keys.
- Keep Deep Analysis jobs, events, artifacts, and proposals owner-scoped.
- Use short-lived job-scoped credentials for Python-to-Go provider tools; never
  expose those credentials in logs, browser responses, or replayable SSE.
- Keep provider URL/citation allowlists intact.
- Persist only bounded public provider event summaries; full claims remain on
  the internal stream and in confirm-gated reports/proposals.

### Frontend and browser surface

- Treat any AI- or user-supplied rich text as untrusted until sanitized.
- Keep token-handling decisions aligned with the auth design in [authentication.md](authentication.md); any move between `localStorage`, headers or cookies needs a deliberate review of the XSS and CSRF trade-offs.
- Avoid leaking sensitive identifiers through URLs, browser history, or cacheable responses where a request body or header would do.

### Supply chain and delivery

- Keep build inputs reproducible: pinned actions, reviewed dependency bumps, and explicit scanner config.
- Generated docs may contain fake examples; scanners should allowlist those precise artifacts rather than globally weakening rules.
- Treat secret-scan failures and high-risk dependency alerts as blockers; dependency dashboards are maintenance inputs, not optional chores.
- Temporary scanner exceptions must be narrow and carry an owner, expiration date, and linked threat-model or issue reference.

## How to add or change a principle

1. Update this document with the control language to follow in practice, and explain the change in the pull request.
2. If the change introduces a new risk or retires an old one, update [threat-model.md](threat-model.md) in the same PR.
3. If the change affects incident handling, update [incident-response.md](incident-response.md) and [SECURITY.md](../SECURITY.md) together.

## Related documents

- [threat-model.md](threat-model.md)
- [incident-response.md](incident-response.md)
- [authentication.md](authentication.md)
- [references.md](references.md)
- [deployment.md](deployment.md)
- [SECURITY.md](../SECURITY.md)
