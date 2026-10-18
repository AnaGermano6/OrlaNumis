# Threat Model

> The live inventory of known security threats and security-review findings for Orla Numis. This document keeps the original **backend / frontend / supply-chain** split because those are the operational owners and remediation surfaces in this repo.

## How to add a new threat finding

1. Add the finding to the correct surface area below (`B-*`, `F-*`, or `SC-*`).
2. Record **severity**, **status**, **location**, **description**, and **recommended remediation** in the same edit.
3. Use `Open`, `Mitigated`, `Accepted` or `Retired` (the feature was removed) as the status. If prioritization matters, append it in parentheses (for example `Open (P1)`).
4. Link any architecture or policy change needed for remediation to [security-principles.md](security-principles.md).
5. If the finding becomes an active incident, manage the operational response in [incident-response.md](incident-response.md) and keep this document focused on the durable risk record.

## Status summary

| Domain | Findings | Mitigated | Open | Accepted | Retired |
|---|---:|---:|---:|---:|---:|
| Backend API | 14 | 8 | 1 | 0 | 5 |
| Frontend | 8 | 3 | 3 | 1 | 1 |
| Supply chain & infrastructure | 7 | 4 | 2 | 0 | 1 |
| **Total** | **29 enumerated** | **15** | **6** | **1** | **7** |

**Last reconciliation:** 2026-10-17. Passkeys (WebAuthn), API keys, the external tool server, the MCP adapter, OIDC sign-in and the Docker publishing workflows were removed, so B-7, B-8, B-10, B-14, F-7 and SC-1 are retired, like B-11 before them. Earlier mitigations include B-2 (SQL injection whitelist), B-6 (request size limits), B-12 (Deep Analysis owner scoping, job credentials, provider allowlists, bounded public events and confirm-gated writes), B-13 (Coin Copilot owner scoping, bounded execution, replay privacy and read-only execution credentials), F-1/F-2/F-4 (DOMPurify sanitization for XSS) and SC-2 (Taskfile secret generation). The table reflects the current code.

## Backend API findings

| ID | Severity | Status | Location | Description | Recommended remediation |
|---|---|---|---|---|---|
| B-1 | Critical | Mitigated | `src/api/main.go`, `src/api/config/config.go` | CORS previously reflected arbitrary origins while allowing credentials. | Keep `CORS_ORIGINS` allowlisted and review any new origin-handling logic carefully. |
| B-2 | Critical | Mitigated | `src/api/handlers/analysis.go` | A user-controlled column name was previously concatenated into a query, but is now protected by an explicit whitelist map in `DeleteAnalysis()` (lines 229–238) and switch-based validation in `Analyze()` (lines 175–185). | Maintain the whitelist pattern for any future dynamic column handling. |
| B-3 | High | Mitigated | `src/api/config/config.go` | The application previously allowed a weak default JWT secret. | Preserve the startup fail-fast and minimum-length checks; never weaken them for convenience. |
| B-4 | High | Mitigated | `src/api/handlers/images.go` | Uploads now enforce an extension allowlist, but the original finding still has a residual MIME-validation gap. | Add magic-byte validation so uploads are checked by content, not extension alone. |
| B-5 | High | Mitigated | `src/api/main.go` | Auth endpoints previously allowed unlimited brute-force attempts. | Keep auth rate limiting in place and extend it to new auth-adjacent endpoints when added. |
| B-6 | Medium | Mitigated | `src/api/main.go` (line ~130: `r.MaxMultipartMemory = middleware.DefaultMultipartMemoryBytes`), middleware | Request bodies and multipart uploads now have explicit size caps enforced by Gin. | Maintain the configured memory limit; document enforcement in deployment documentation. See issue #201. |
| B-7 | Medium | Retired | — | In-memory WebAuthn ceremony sessions needed a TTL. | Passkeys (WebAuthn) were removed (2026-10); username and password is the only sign-in method. |
| B-8 | Medium | Retired | — | WebAuthn origin validation trusted request headers. | Passkeys (WebAuthn) were removed (2026-10). |
| B-9 | Low | Open (P3) | `src/api/handlers/numista.go` | Some error responses expose more internal detail than clients need. | Return generic client-facing errors and keep specifics in logs only. [#163](https://github.com/briandenicola/coin-collection-app/issues/163) |
| B-10 | High | Retired | — | The external tool server exposed collection data to API-key clients. | The external tool server (`/api/v1/tools/*`) and API keys were removed (2026-10). |
| B-11 | High | Retired | — | NumisBids and CNG provider credentials were stored per-user. | The auction-provider, ParcelApp and Pushover integrations were removed (2026-10), with their stored credentials and `AUCTION_CREDENTIAL_ENCRYPTION_KEY`; no provider secret is stored any more. |
| B-12 | High | Mitigated | `src/api/services/deep_identification_service.go`, `src/api/services/deep_identification_pipeline_runner.go`, `src/api/handlers/internal_tools.go`, `src/api/middleware/internal_token.go`, `src/agent/app/teams/deep_identification/` | Deep Analysis introduces persisted background jobs, uploaded hint artifacts, replayable SSE, and Python-to-Go provider callbacks. Mitigations include owner-scoped repositories, ephemeral hint cleanup, short-lived job-scoped callback tokens, provider/citation allowlists, fixed-template OCRE SPARQL, bounded public event payloads, and explicit proposal review/apply before any coin write. | Preserve all layers together. New providers must add explicit network/license validation, typed bounded evidence, failure isolation, and citation allowlisting. Never place internal tokens or full raw provider payloads in replayable events. |
| B-13 | High | Mitigated | `src/api/services/coin_copilot_*`, `src/api/handlers/coin_copilot*`, `src/api/handlers/internal_tools.go`, `src/agent/app/teams/coin_copilot.py` | Coin Copilot introduces durable prompts, checkpoints, replayable SSE, resumable execution, and Python-to-Go callbacks. Mitigations are default-off rollout with legacy fallback, fail-closed model capability preflight, server-derived owner scope, read-only allowlisted tools, fresh owner/run/execution-bound credentials, transactional cancellation and event sequencing, 120-second default/150-second maximum execution timeout within the absolute 180-second token TTL, bounded payloads and retention, and rejection of chain-of-thought/provider-native traces. Dollar-cost enforcement is deferred, but reliable token usage stays observable and iteration/tool/time/concurrency/payload limits remain enforced. | Preserve Go ownership of data, retention, cancellation, and authorization. Keep Python stateless/DB-free; never add write, web, shell, filesystem, or arbitrary HTTP tools to this token family. Keep prompts, collection content, raw tool data, credentials, and hidden reasoning out of logs/events. Roll back by disabling `CoinCopilotEnabled` while retaining read/cancel access for existing runs. |
| B-14 | High | Retired | — | The native MCP adapter exposed collection data and Coin Copilot lifecycle tools to API-key clients. | The MCP adapter (`/api/mcp`) and API keys were removed (2026-10). |

## Frontend findings

| ID | Severity | Status | Location | Description | Recommended remediation |
|---|---|---|---|---|---|
| F-1 | Critical | Mitigated | `src/web/src/components/coin/CoinAIAnalysis.vue` | AI analysis Markdown was rendered unsanitized, but is now passed through `DOMPurify.sanitize()` before injection (lines 33–35). | Keep DOMPurify pinned to a stable version and monitor upstream security advisories. |
| F-2 | Critical | Mitigated | `src/web/src/composables/useCoinSearchChat.ts` | Chat messages were transformed and injected without sanitization, but `formatMessage()` now sanitizes all HTML via `DOMPurify.sanitize()` before rendering in `CoinSearchChat.vue` line 31. | Ensure all HTML rendering paths go through the sanitization function. |
| F-3 | Critical | Open (P3) | `src/web/src/stores/auth.ts` | JWT and refresh tokens live in `localStorage`, which turns any XSS into token theft. | Short term: eliminate XSS sinks (F-1/F-2/F-4 mitigations complete the path). Long term: evaluate `HttpOnly` cookie transport. [#163](https://github.com/briandenicola/coin-collection-app/issues/163) |
| F-4 | High | Mitigated | `src/web/package.json` (DOMPurify ^3.4.1), `src/web/src/composables/useCoinSearchChat.ts`, `src/web/src/components/coin/CoinAIAnalysis.vue` | Rich-text rendering is now backed by DOMPurify (`@types/dompurify` ^3.2.0) and sanitized at all HTML injection points. | Keep DOMPurify pinned and monitor for updates; audit any new HTML rendering paths to confirm sanitization. See [security-principles.md](security-principles.md). |
| F-5 | High | Open (P3) | `src/web/src/api/client.ts` | Refresh tokens travel in a JSON body rather than an `HttpOnly` cookie-based flow. | Revisit transport if auth is redesigned; document the trade-off in [authentication.md](authentication.md). [#163](https://github.com/briandenicola/coin-collection-app/issues/163) |
| F-6 | Medium | Open (P2) | frontend auth responses (Go API: `src/api/handlers/auth.go`; Vue client proxies responses) | Sensitive responses are not explicitly marked `Cache-Control: no-store`. | Set no-store headers on auth and token-bearing responses at the Go API level (handlers) and confirm Vue client respects them. [#163](https://github.com/briandenicola/coin-collection-app/issues/163) |
| F-7 | Medium | Retired | — | The WebAuthn login put the username in a URL. | Passkeys (WebAuthn) were removed (2026-10). |
| F-8 | Low | Accepted (platform limitation) | `src/web/src/pages/LoginPage.vue` | Password state cannot be meaningfully scrubbed from browser memory after use in a reliable, portable way. | Accept the platform limit; keep lifetime short in reactive state and avoid unnecessary copies. |

## Supply chain and infrastructure findings

| ID | Severity | Status | Location | Description | Recommended remediation |
|---|---|---|---|---|---|
| SC-1 | Critical | Retired | — | The Docker publishing workflows used mutable action references. | The Docker publishing workflows were removed (2026-10); the remaining workflows pin `uses:` to commit SHAs. |
| SC-2 | Critical | Mitigated | `Taskfile.yml`, `src/api/config/config.go` | JWT secret was previously hardcoded, but is now injected via environment variable at runtime. The `gen-env` task generates a random secret and stores it in a `.env` file (Taskfile.yml lines 143–145). Config enforces minimum 32-character length and fails fast if unset (config.go lines 21–33). | Maintain the env-based pattern; document that production deployments must set `JWT_SECRET` in CI/CD secrets before container start. |
| SC-3 | Critical | Mitigated | `src/web/scripts/download-background-removal-assets.mjs`, `src/web/scripts/background-removal-assets.lock.json` | The build downloads only the required `@imgly/background-removal` 1.7.0 chunks from a fixed origin. A committed content-addressed lockfile pins every chunk; the downloader rejects version/origin/path drift, redirects, size mismatches, and SHA-256 mismatches before atomically publishing the assets. | Review and regenerate the lockfile deliberately whenever `@imgly/background-removal` is upgraded. |
| SC-4 | High | Open (P1) | `src/api/go.mod` | `golang.org/x/*` dependencies were called out for lagging current versions in the original review. | Re-check with `govulncheck` / dependency review and upgrade deliberately. [#163](https://github.com/briandenicola/coin-collection-app/issues/163) |
| SC-5 | High | Open (P2) | GitHub repository settings | Branch protection is not documented as enforced for `main`/`beta`. | Require PR reviews, required checks, and restricted direct pushes in GitHub Settings → Branches. [#163](https://github.com/briandenicola/coin-collection-app/issues/163) |
| SC-6 | Medium | Mitigated | `Dockerfile`, `src/agent/Dockerfile` | Production base images use reviewed tag-plus-OCI-index-digest references, including the Go API builder on the fixed Go 1.26.6 patch line. | Keep the tag and digest paired; refresh monthly or when a base-image CVE requires it. [#320](https://github.com/briandenicola/coin-collection-app/issues/320) |
| SC-7 | Medium | Mitigated | `Dockerfile`, `src/agent/Dockerfile`, `docs/deployment.md` | Final application containers now run as non-root UID/GID `10001:10001`, with runtime-owned app, data, uploads, and agent paths. | Keep persistent bind mounts writable by UID/GID `10001:10001`; see `docs/deployment.md`. [#319](https://github.com/briandenicola/coin-collection-app/issues/319) |

## Related documents

- [security-principles.md](security-principles.md)
- [incident-response.md](incident-response.md)
- [authentication.md](authentication.md)
- [SECURITY.md](../SECURITY.md)
