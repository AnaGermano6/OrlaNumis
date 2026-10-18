# Testing Strategy

This document explains what we test, what we intentionally do not test, and how to add tests across the Go API, the Vue PWA and the Python agent.

## 1. Testing Philosophy

**Confidence over coverage.** We optimize for regressions affecting users, data integrity and architecture, not percentages. Every new service method gets at least one unit test, and a bug fix gets a regression test for the exact failing path.

**Fast feedback first.** Run targeted tests while working, then the full gates (§5) before calling a change done.

**Test the contract, not the implementation.** Handler tests exercise HTTP status codes and JSON payloads, store tests assert state transitions, and agent tests validate schemas, routing helpers and endpoint contracts instead of private internals.

**Architecture tests are not optional.** `src/api/architecture_test.go` enforces the layer rules; passing feature tests cannot compensate for a failing architectural guard.

**Integration evidence follows risk.** Deterministic cross-service seams and DB-backed tests catch failures isolated unit mocks cannot. Model-quality evaluation stays separate from deterministic orchestration tests.

## 2. Test Surface Inventory

| Service | Layer | Tool | Location | Run command | What it tests |
|---|---|---|---|---|---|
| Go API | Architecture | `go test` | `src/api/architecture_test.go` | `cd src/api && go test -v -run "TestNoDirectDatabaseImports|TestHandlersDoNotUseRawSQL|TestPackageImportMatrix" .` | 1 file / 3 tests enforcing DI-only database access, no raw SQL in handlers, and the package import matrix. |
| Go API | Unit + package-level behavior | `go test` | `src/api/{handlers,middleware,repository,services}/*_test.go` | `cd src/api && go test -v ./...` | Handler, middleware, repository, and service behavior using HTTP test routers and in-memory SQLite. |
| Go API | Dedicated integration / E2E | `go test` | `src/api/integration/*_test.go` | `cd src/api && go test -v ./integration/...` | Full DB-backed service/handler workflow tests (Numista compatibility/performance/security/workflows). Not a cross-process suite by default. |
| Go API | Go↔Python seam (cross-process) | `go test -tags=seam` | `src/api/integration/deep_identification_seam_test.go` | See §7 below | Boots the real Python agent service and drives the real `DeepIdentificationPipelineRunner` over a real HTTP/SSE round trip. Excluded from CI by build tag + env var. |
| Vue Web | Type checking | `vue-tsc` | `src/web/package.json`, `src/web/src/**` | `cd src/web && npm run type-check` | CI runs the package script (`vue-tsc --build`); `--noEmit` is not a substitute. |
| Vue Web | Lint | ESLint | `src/web/package.json`, `src/web/eslint.config.*` | `cd src/web && npm run lint` | CI runs zero-warning lint (`--max-warnings 0`); never replace it with `--quiet`. |
| Vue Web | Unit / component / store / API tests | Node test runner + Vitest | `src/web/scripts/`, `src/web/src/**/__tests__/*` | `cd src/web && npm run test` | The script runs asset-script tests plus Vitest; calling Vitest alone omits part of the gate. |
| Vue Web | Build parity | `vite` + `vue-tsc --build` | `src/web/package.json` | `cd src/web && npm run build` | Production build gate. This is stricter than `vue-tsc --noEmit`; nullable props and indexed access that pass locally can still fail here. |
| Vue Web | AI browser exploration | Vitest + Playwright | `src/web/e2e/exploration/` | `npm run test` (its contract tests); runs through `docker-compose.exploration.yml` and `.github/workflows/ai-browser-exploration.yml` | Contract, budget and orchestration tests of the AI browser explorer. The exploration itself drives an isolated stack with a model and is not part of the Quality Gate. |
| Python Agent | Lint | Ruff | `src/agent/app/`, `src/agent/tests/`, `src/agent/pyproject.toml` | See §5 (`task check:agent`) | Import order, correctness, and style rules for the deterministic Python surface. |
| Python Agent | Unit / contract tests | pytest | `src/agent/tests/test_*.py` | See §5 (`task check:agent`) | Request validation, schemas, provider/routing behavior, streaming, and deterministic orchestration. |
| Python Agent | Static type checking | None configured today | `src/agent/pyproject.toml` | N/A | No `mypy` or `pyright` config is present; schema validation and pytest carry the current contract burden. |

Notes:
- `src/api/integration/` contains a dedicated integration package in addition to package-local HTTP tests.
- Cross-service contracts shared by the Go API, the Python agent and the web app (Coin Copilot hand-off fixtures, browser exploration schemas) live in [`contracts/`](../contracts/).
- Test counts change; use current runner output as evidence, not historical counts.

## 3. What we DON'T test

- We do **not** test SQLite, GORM, Gin, Vitest, FastAPI or LangGraph themselves; we test our usage of them.
- We do **not** hit third-party services in automated tests. Anthropic, Ollama, SearXNG, Numista, the ECB and the mints' shops are mocked or stubbed at the service boundary.
- Scraper tests use sanitized fixtures under `src/api/services/testdata/` or `httptest` stubs. HAR captures and browser exports from real sessions are sensitive (cookies, headers, account identifiers); `*.har` is ignored by `.gitignore` and must not be committed.
- We do **not** attempt deterministic end-to-end assertions for LLM quality. For agent work we test input validation, parsing, retry behavior, routing helpers and schema-shaped outputs.
- We do **not** add tests for trivial getters or thin pass-through code just to raise coverage.

Deep Analysis follows these rules across all three services. Python tests exercise the router override, bounded fan-out, typed provider failures, deterministic disagreement detection and Pydantic stream frames without live LLM or provider calls. Go `httptest` streams feed those Python-shaped frames through the real pipeline runner to verify public SSE translation, persisted proposals, owner-scoped lifecycle, provider telemetry and hint cleanup. Vue component tests verify reconnect, exact coverage states and confirm-gated draft/coin writes.

## 4. Adding a New Test

### Go API

- Put the test beside the package under test as `*_test.go`; do not create a separate test tree.
- Follow the helper pattern in `src/api/services/auth_service_test.go` and `src/api/handlers/auth_handler_test.go`: small setup functions marked with `t.Helper()`.
- Use real in-memory SQLite (`glebarez/sqlite` + `gorm.Open(...":memory:"...)`) when repository behavior, transactions, or auth persistence matter.
- Use `httptest` + Gin routers for handler and middleware tests; assert status codes, JSON shape, and authorization behavior.
- Prefer table-driven tests for pure logic and parser-style code; `src/api/services/valuation_parser_test.go` is the model.
- Mock only true external boundaries (agent proxy, networked services). Prefer real repos/services inside the process.
- If a change touches the layer rules, extend `src/api/architecture_test.go` rather than burying the rule in prose.

### Vue Web

- Put tests under a nearby `__tests__/` folder, as in `src/web/src/components/__tests__/CollectionPagination.test.ts` or `src/web/src/stores/__tests__/auth.test.ts`.
- Test public behavior: rendered text, emitted events, store state, request payloads, or design-token guarantees—not component internals.
- Use Vitest plus Vue Test Utils for mounted components and `vi.mock()` for boundary modules such as `@/api/client`.
- Stub browser globals explicitly with `vi.stubGlobal`, as shown in `src/web/src/api/__tests__/client.test.ts` and `src/web/src/stores/__tests__/auth.test.ts`.
- Keep fixtures small and inline until multiple files need the same data.
- Source-scanning tests are acceptable for structural UI rules when rendering is unnecessary; see `src/web/src/__tests__/design-tokens.test.ts` and `src/web/src/pages/__tests__/CollectionPage.test.ts`.
- Finish with `task check:web`: zero-warning lint, strict type-check, the full package tests and the production build.

### Python Agent

- Put tests in `src/agent/tests/` as `test_*.py`.
- Use pytest fixtures to remove nondeterminism or waiting; `src/agent/tests/test_retry.py` patches retry delays to zero.
- Use FastAPI `TestClient` for route-level contracts, as in `src/agent/tests/test_api.py`.
- Stub LLM/model calls with `AsyncMock` or direct helper invocation rather than making live provider calls.
- Favor testing deterministic helpers directly—`parse_verdicts` in `src/agent/tests/test_availability.py` and `_build_coin_show_location_context` in `src/agent/tests/test_supervisor_coin_shows_location.py` are good examples.
- Assert structured outputs and schema defaults, not prose quality.
- Keep provider/network behavior at the boundary; real Anthropic/Ollama/SearXNG traffic does not belong in CI.

## 5. Running the Gates (local and CI)

The [Taskfile](../Taskfile.yml) is the shared entry point, and the Quality Gate ([`ci.yml`](../.github/workflows/ci.yml)) runs the same targets on every pull request to `main`:

| Target | What it runs |
|---|---|
| `task check:go` | `go build ./...`, `go vet ./...`, `gofmt -l .` (fails on any unformatted file) and `go test -v ./...` in `src/api` |
| `task test-race` | `go test -race ./...` with `CGO_ENABLED=1` (a separate CI job) |
| `task check:openapi` | Regenerates the Swagger artifacts and `docs/openapi.json` (`task openapi`) and fails if they differ from the committed ones; **writes files** |
| `task check:web` | `npm run lint` (zero warnings), `npm run type-check` (`vue-tsc --build`), `npm run test` (asset-script tests plus Vitest) and `npm run build` in `src/web` |
| `task check:agent` | Verifies the locked environment (`uv sync --locked --check --offline --extra dev`), then Ruff and the full pytest suite |
| `task check` | `check:go`, `check:web`, `check:agent` and `check:openapi` in turn |

Preparation is separate: `task setup:go`, `task setup:web` (`npm ci`), `task setup:agent` (`uv sync --locked --extra dev`) and `task setup:openapi` (swag v1.16.6). CI uses Task 3.44.0, Node 24.15.0, Go from `src/api/go.mod`, and Python 3.12 with uv 0.11.22.

Go commands run with `GOTOOLCHAIN=local` and `GOPROXY=off`, so a missing setup fails instead of downloading. After changing an API annotation, run `task openapi` and commit the regenerated files.

On Windows the Taskfile uses `npm.cmd`; do the same for direct package commands.

## 6. Golden collection fixtures

Shared "golden" coin fixtures give the Go and web tests the same deterministic data:

| Fixture | Traits |
|---|---|
| `roman-denarius-core` | roman |
| `greek-tetradrachm-valued` | greek, valued |
| `byzantine-solidus-set-member` | byzantine, set-member |
| `wishlist-aureus-target` | roman, wishlist, valued |
| `legacy-era-provincial-bronze` | roman, legacy custom era |
| `diocletian-follis` | roman |
| `image-heavy-drachm` | greek, image-heavy |
| `reference-rich-denarius` | roman, reference-rich, set-member |

The traits describe the fixture coins, not the app's categories: the fixtures use the `Modern` and `Other` categories.

- **Go** — import `src/api/testutil`; use `BuildGoldenCoinFixture(name, userID)`, a named helper such as `BuildDiocletianFollis(userID)`, or `BuildGoldenCoinFixtures(userID)`. `PersistGoldenCollection(db, userID)` seeds the coins with their sets, set memberships, images and references into a migrated test database. `src/api/testutil/coin_fixtures_test.go` checks the traits, clone safety and persisted associations.
- **Web** — import from `src/web/src/test/fixtures`; use `buildGoldenCoinFixture(name)`, a named helper such as `buildImageHeavyDrachm()`, or `buildGoldenCoinFixtures()`, and `buildTestCoinSets()` / `buildTestMintLocations()` for related data. `src/web/src/test/fixtures/coins.test.ts` checks the traits, clone safety and key associations.

## 7. Go↔Python deep-identification seam test

Every deep-identification test elsewhere in this repo (`src/api/services/deep_identification_pipeline_runner_stream_test.go` and the Python `tests/` suite) drives its own side against hand-written, convention-only fixtures for the *other* side's wire shape. That is exactly the gap behind the 080e598 production outage: both sides were internally consistent and both were wrong about each other. `src/api/integration/deep_identification_seam_test.go` closes that gap by booting the **real** `uvicorn app.main:app` Python process and driving the **real**, exported `DeepIdentificationPipelineRunner.Run` (over the real `AgentProxy.StreamDeepIdentification` HTTP/SSE client) against it — no fixture on either side.

### Why it is excluded from unattended CI

The test needs a real Python interpreter/venv and takes real wall-clock time (worker startup + the LLM retry/backoff ladder), so it is guarded twice:

1. **Build tag**: the file starts with `//go:build seam`. `go build ./...`, `go vet ./...`, and `go test ./...` never see or compile it without `-tags=seam`.
2. **Env var**: even built with the tag, the test calls `t.Skip` unless `DEEP_SEAM_TEST=1` is set.

Both are required to actually run it — this is deliberate defense in depth.

### How to run it

Prerequisites:
- `src/agent/.venv` must exist and be able to import `uvicorn`, `langchain_ollama`, and `langchain_anthropic` (run `pip install -e ".[dev]"` from `src/agent` first if not).
- No live LLM credentials or network access are required — see the LLM tradeoff below.

```powershell
cd src/api
$env:DEEP_SEAM_TEST = "1"
go test -tags=seam -run TestDeepIdentificationSeam -v ./integration/...
```

On Linux/macOS:

```bash
cd src/api
DEEP_SEAM_TEST=1 go test -tags=seam -run TestDeepIdentificationSeam -v ./integration/...
```

If your Python interpreter is not at `src/agent/.venv/Scripts/python.exe` (Windows) or `src/agent/.venv/bin/python` (POSIX), point `DEEP_SEAM_PYTHON` at it.

**Expected runtime**: roughly 10-20 seconds (Python process startup plus the vision node's LLM retry/backoff ladder against a deliberately unreachable endpoint — see below). If the Python service fails to become healthy within 30 seconds, or the venv/interpreter cannot be found, the test skips or fails with the captured stdout/stderr from the agent process.

### The LLM tradeoff (documented, not hidden)

The test configures the LLM provider as Ollama pointed at a local TCP port nothing is listening on. This is **not** a stub of the seam: FR-006/FR-040 already require every LLM call site in the deep-identification pipeline (vision hypothesis, evaluator disagreement summary, synthesis narrative) to degrade to a deterministic fallback on any LLM failure, never to raise. Pointing at an unreachable endpoint exercises a real call that fails fast and is handled by the pipeline's own documented resilience path — so the test runs unattended with no API key, no external network egress, and no nondeterministic model output, while still genuinely exercising the vision node's real structured-output call path.

Similarly, the provider catalog is left at its real production default, but `tools_base_url` is left empty — the exact same code path production uses when the tools client is unset — so `numista` settles immediately as `unconfigured` with zero upstream calls, while `ngc`/`rpc` (never automated) still run their real, always-network-free provider nodes. No `numista.com` network traffic occurs.


## 8. Coin Copilot contract and operations testing

Coin Copilot uses deterministic seam tests instead of live-model assertions.
Required coverage spans:

- Go state transitions, owner scoping, idempotent start/resume, cancellation
  races, stale recovery, retention, bounded payloads, token usage, and exactly
  one terminal event;
- execution-credential signature, expiry, revocation, owner/run/execution/tool
  binding, and rejection of all write/arbitrary tools;
- Python strict schemas, model capability preflight, malformed calls,
  prompt-injection-as-data, iteration/tool/time exhaustion, sequential tool
  execution, payload truncation, and provider-reported token accounting;
- Vue feature-off/unsupported-model legacy fallback, progress,
  clarification/resume, cancellation, reconnect/de-duplication, and
  desktop/PWA rendering;
- route/OpenAPI and shared Go/Python fixture drift.

The timeout matrix must cover the 120-second default and 150-second maximum.
Values above 150 are invalid because the execution timeout plus the 30-second
credential buffer must not exceed the token's absolute 180-second TTL.
Dollar-cost enforcement is intentionally absent; tests should assert that
reliable input/output token counts remain observable and that iteration,
tool-call, wall-clock, sequential-concurrency, and payload limits remain
enforced.

Use mocked provider responses and fake internal SSE streams; do not contact
Anthropic, Ollama, or other external services. The shared Go/Python fixtures of
the hand-off to Deep Analysis live in
[`contracts/coin-copilot/`](../contracts/coin-copilot/).

Documentation/API drift verification:

```powershell
task openapi
git diff --exit-code -- src\api\docs\docs.go src\api\docs\swagger.json src\api\docs\swagger.yaml docs\openapi.json
Push-Location src\api
go test . -run TestRegisteredAPIRoutesAreDocumentedInOpenAPI -count=1
Pop-Location
```

Feature-specific suites, before the full gates:

```powershell
Push-Location src\api
go test ./handlers ./repository ./services -run CoinCopilot -count=1
Pop-Location

Push-Location src\agent
pytest tests/test_coin_copilot_contract.py tests/test_coin_copilot_capabilities.py tests/test_coin_copilot_security.py tests/test_coin_copilot_harness.py -v
Pop-Location

Push-Location src\web
npx vitest run src/composables/__tests__/useCoinCopilot.test.ts src/components/__tests__/CoinSearchChat.test.ts src/components/admin/__tests__/AdminSystemSection.coin-copilot.test.ts
Pop-Location
```

These commands document the expected validation path; passing results must not
be claimed unless the commands were actually executed.

## 9. Cross-references

- System architecture: [`ARCHITECTURE.md`](ARCHITECTURE.md)
- Go API README: [`../src/api/README.md`](../src/api/README.md)
- Vue frontend README: [`../src/web/README.md`](../src/web/README.md)
- Python agent README: [`../src/agent/README.md`](../src/agent/README.md)
- Quality Gate workflow: [`../.github/workflows/ci.yml`](../.github/workflows/ci.yml)
- Task runner: [`../Taskfile.yml`](../Taskfile.yml)

Python static type checking (mypy or pyright) is not configured; schema validation and pytest carry the contract burden.
