# API Reference

Every endpoint of the Orla Numis API, grouped by area. It is generated from [`openapi.json`](openapi.json), the OpenAPI document produced from the code's annotations (`task openapi`), which also has each endpoint's parameters, bodies and responses. A running app serves it interactively at `/swagger/index.html`.

- **Base path** — `/api` (for example `http://localhost:8080/api/coins`).
- **Authentication** — every endpoint except `/auth/login`, `/auth/register` and `/auth/refresh` needs `Authorization: Bearer <access token>`. The container liveness checks (`/health`, `/healthz`) are outside `/api`. Get a token from `POST /api/auth/login` and renew it with `POST /api/auth/refresh`. See [Authentication](authentication.md).
- **Ownership** — data endpoints only reach the signed-in user's own coins, sets and drafts. `/admin/*` endpoints need the admin role.
- **Errors** — failures answer with a JSON body `{"error": "…"}` and the matching HTTP status.
- **Internal** — `/internal/*` endpoints are called by the Python agent with its service token, never by browsers.

To regenerate this page after changing the API:

```sh
task openapi
```

and update the tables below from `docs/openapi.json`.


## Auth

| Method | Path | Summary |
|---|---|---|
| `POST` | `/auth/login` | Login |
| `POST` | `/auth/refresh` | Refresh tokens |
| `POST` | `/auth/register` | Register a new user |

## User

| Method | Path | Summary |
|---|---|---|
| `POST` | `/auth/change-password` | Change password |
| `GET` | `/auth/me` | Get current user |
| `POST` | `/user/avatar` | Upload avatar |
| `DELETE` | `/user/avatar` | Delete avatar |
| `GET` | `/user/export` | Export collection |
| `GET` | `/user/export/catalog` | Export PDF catalog |
| `POST` | `/user/import` | Import collection |
| `PUT` | `/user/profile` | Update profile |

## Coins

| Method | Path | Summary |
|---|---|---|
| `GET` | `/coin-catalogue` | A country's coin catalogue |
| `GET` | `/coin-catalogue/countries` | Countries of the coin catalogue |
| `GET` | `/coin-catalogue/joint-issues` | Joint issues of the euro area |
| `PUT` | `/coin-catalogue/overrides` | Hide a coin of the catalogue, or join it to another |
| `DELETE` | `/coin-catalogue/overrides` | Show a coin of the catalogue again |
| `GET` | `/coin-options` | Get coin category and era options |
| `GET` | `/coin-photos/search` | Search real coin photos |
| `GET` | `/coin-releases` | List new and upcoming coins |
| `GET` | `/coin-releases/numista-type` | Find the Numista type of a new coin |
| `GET` | `/coins` | List coins |
| `POST` | `/coins` | Create a coin |
| `POST` | `/coins/bulk` | Bulk coin action |
| `GET` | `/coins/export` | Export coins as a spreadsheet |
| `GET` | `/coins/countries` | List collection countries |
| `POST` | `/coins/country-from-numista` | Fill coin countries from Numista |
| `GET` | `/coins/denominations` | List collection face values |
| `POST` | `/coins/intake/commit` | Commit coin intake draft |
| `POST` | `/coins/intake/draft` | Create coin intake draft |
| `POST` | `/coins/lookup` | Coin lookup from images |
| `POST` | `/coins/match-category-era` | Match a category/era candidate against known values |
| `GET` | `/coins/numista-fill` | Preview a fill of coin records from Numista |
| `POST` | `/coins/numista-fill` | Fill coin records from Numista |
| `POST` | `/coins/share-list` | Download a list to share |
| `GET` | `/coins/statuses` | List coins by legal tender and commemorative status |
| `GET` | `/coins/summary` | Get collection summary |
| `GET` | `/coins/{id}` | Get a coin |
| `PUT` | `/coins/{id}` | Update a coin |
| `DELETE` | `/coins/{id}` | Delete a coin |
| `POST` | `/coins/{id}/duplicate` | Duplicate a coin |
| `POST` | `/coins/{id}/images/import` | Import a real coin photo |
| `GET` | `/coins/{id}/market-value` | Get a coin's Numista market value |
| `POST` | `/coins/{id}/official-photos` | Add a coin's official images |
| `POST` | `/coins/{id}/purchase` | Mark coin as purchased |
| `GET` | `/coins/{id}/recommendations` | List coin recommendations |
| `POST` | `/coins/{id}/recommendations/{recommendationId}/accept` | Accept recommendation |
| `POST` | `/coins/{id}/recommendations/{recommendationId}/reject` | Reject recommendation |
| `GET` | `/coins/{id}/status` | Get a coin's legal tender and commemorative status |
| `GET` | `/coins/{id}/value-history` | Get coin value history |
| `GET` | `/coins/{id}/wishlist-matches` | Find a collection coin on the wishlist |
| `GET` | `/exchange-rates` | Get exchange rates |
| `GET` | `/stats` | Get collection statistics |
| `GET` | `/stats/distribution` | Get collection distribution |
| `GET` | `/suggestions` | Get autocomplete suggestions |

## Images

| Method | Path | Summary |
|---|---|---|
| `POST` | `/coins/{id}/images` | Upload a coin image |
| `POST` | `/coins/{id}/images/swap-sides` | Swap a coin's obverse and reverse images |
| `DELETE` | `/coins/{id}/images/{imageId}` | Delete a coin image |
| `GET` | `/proxy-image` | Proxy an external image |
| `GET` | `/scrape-image` | Scrape image URL from a web page |
| `GET` | `/uploads/{filepath}` | Serve authorized uploaded media |

## Coin References

| Method | Path | Summary |
|---|---|---|
| `POST` | `/coins/{id}/references` | Create coin reference |
| `PUT` | `/coins/{id}/references/{referenceId}` | Update coin reference |
| `DELETE` | `/coins/{id}/references/{referenceId}` | Delete coin reference |
| `POST` | `/references/migrate-legacy` | Migrate legacy references |

## Journal

| Method | Path | Summary |
|---|---|---|
| `GET` | `/coins/{id}/journal` | List journal entries |
| `POST` | `/coins/{id}/journal` | Add journal entry |
| `DELETE` | `/coins/{id}/journal/{entryId}` | Delete journal entry |

## Analysis

| Method | Path | Summary |
|---|---|---|
| `GET` | `/ai-status` | Check AI provider status |
| `POST` | `/coins/{id}/analyze` | Queue coin image analysis |
| `DELETE` | `/coins/{id}/analyze` | Delete coin analysis |
| `POST` | `/coins/{id}/grade` | Queue coin grading |
| `POST` | `/extract-text` | Extract text from image |
| `GET` | `/ollama-status` | Check Ollama status |

## AI Jobs

| Method | Path | Summary |
|---|---|---|
| `GET` | `/ai-jobs/{id}` | Get AI job |
| `GET` | `/coins/{id}/ai-jobs` | List coin AI jobs |

## Wishlist

| Method | Path | Summary |
|---|---|---|
| `PUT` | `/coins/{id}/listing-status` | Update coin listing status |
| `GET` | `/wishlist/availability-runs` | List my wishlist availability runs |
| `GET` | `/wishlist/availability-runs/{id}` | Get my wishlist availability run detail |
| `POST` | `/wishlist/check-availability` | Check wishlist availability |
| `POST` | `/wishlist/url-intake/analyze` | Analyze wishlist listing URL |

## Wishlist Search Alerts

| Method | Path | Summary |
|---|---|---|
| `GET` | `/wishlist/search-alerts` | List wishlist search alerts |
| `POST` | `/wishlist/search-alerts` | Create wishlist search alert |
| `PUT` | `/wishlist/search-alerts/{alertId}` | Update wishlist search alert |
| `DELETE` | `/wishlist/search-alerts/{alertId}` | Delete wishlist search alert |
| `GET` | `/wishlist/search-alerts/{alertId}/candidates` | List wishlist search alert candidates |
| `POST` | `/wishlist/search-alerts/{alertId}/candidates/{candidateId}/convert` | Convert wishlist search alert candidate |
| `POST` | `/wishlist/search-alerts/{alertId}/candidates/{candidateId}/dismiss` | Dismiss wishlist search alert candidate |
| `POST` | `/wishlist/search-alerts/{alertId}/candidates/{candidateId}/restore` | Restore wishlist search alert candidate |
| `POST` | `/wishlist/search-alerts/{alertId}/criteria-adjustments` | Adjust wishlist search alert criteria |
| `POST` | `/wishlist/search-alerts/{alertId}/run` | Run wishlist search alert |
| `GET` | `/wishlist/search-alerts/{alertId}/runs` | List wishlist search alert runs |
| `GET` | `/wishlist/search-alerts/{alertId}/runs/{runId}` | Get wishlist search alert run |

## Quick Capture

| Method | Path | Summary |
|---|---|---|
| `GET` | `/quick-capture/drafts` | List Quick Capture drafts |
| `POST` | `/quick-capture/drafts` | Create Quick Capture draft |
| `GET` | `/quick-capture/drafts/{id}` | Get Quick Capture draft |
| `PUT` | `/quick-capture/drafts/{id}` | Update Quick Capture draft |
| `POST` | `/quick-capture/drafts/{id}/discard` | Discard Quick Capture draft |
| `POST` | `/quick-capture/drafts/{id}/promote` | Promote Quick Capture draft |

## Numista

| Method | Path | Summary |
|---|---|---|
| `GET` | `/numista/catalog/issuers` | Search Numista issuers |
| `GET` | `/numista/catalog/search` | Search the Numista coin catalogue |
| `GET` | `/numista/catalog/types/{id}` | Get a Numista coin type |
| `GET` | `/numista/catalog/types/{id}/issues` | List a Numista type's issues |
| `GET` | `/numista/catalog/types/{id}/market-value` | Get a Numista type's market value |
| `POST` | `/numista/enrich` | Enrich and rerank Numista candidates |
| `POST` | `/numista/lookup` | Run broad Numista lookup |
| `POST` | `/numista/query-proposal` | Build a Numista query proposal |
| `GET` | `/numista/search` | Search Numista catalog |
| `GET` | `/translate` | Translate a short text |

## Catalogs

| Method | Path | Summary |
|---|---|---|
| `POST` | `/admin/catalogs` | Create catalog |
| `PUT` | `/admin/catalogs/{id}` | Update catalog |
| `DELETE` | `/admin/catalogs/{id}` | Delete catalog |
| `GET` | `/catalogs` | List catalogs |

## Mint Locations

| Method | Path | Summary |
|---|---|---|
| `POST` | `/admin/mint-locations` | Create mint location |
| `PUT` | `/admin/mint-locations/{id}` | Update mint location |
| `DELETE` | `/admin/mint-locations/{id}` | Delete mint location |
| `GET` | `/mint-locations` | List mint locations |
| `POST` | `/mint-locations` | Create a private mint location |
| `GET` | `/mint-locations/geocode` | Geocode a mint name |
| `PUT` | `/mint-locations/{id}` | Update a private mint location |
| `DELETE` | `/mint-locations/{id}` | Delete a private mint location |

## Sets

| Method | Path | Summary |
|---|---|---|
| `GET` | `/set-builder/proposals/{id}` | Get Agentic set proposal |
| `PUT` | `/set-builder/proposals/{id}` | Update Agentic set proposal |
| `POST` | `/set-builder/proposals/{id}/approve` | Approve Agentic set proposal |
| `POST` | `/set-builder/proposals/{id}/regenerate` | Regenerate Agentic set proposal |
| `POST` | `/set-builder/proposals/{id}/reject` | Reject Agentic set proposal |
| `POST` | `/set-builder/runs` | Queue Agentic set proposal |
| `GET` | `/sets` | List user sets |
| `POST` | `/sets` | Create a set |
| `POST` | `/sets/import-csv` | Create set from CSV |
| `POST` | `/sets/preview-smart` | Preview smart set |
| `GET` | `/sets/{id}` | Get set details |
| `PUT` | `/sets/{id}` | Update a set |
| `DELETE` | `/sets/{id}` | Delete a set |
| `GET` | `/sets/{id}/analytics` | Get set analytics |
| `GET` | `/sets/{id}/checklist` | Get a checklist set's slots |
| `GET` | `/sets/{id}/coins` | List coins in a set |
| `POST` | `/sets/{id}/coins` | Add coin to set |
| `PUT` | `/sets/{id}/coins/order` | Reorder set coins |
| `DELETE` | `/sets/{id}/coins/{coinId}` | Remove coin from set |
| `GET` | `/sets/{id}/completion` | Get set completion |
| `POST` | `/sets/{id}/snapshot` | Create set snapshot |
| `GET` | `/sets/{id}/trends` | Get set trends |

## Stats

| Method | Path | Summary |
|---|---|---|
| `GET` | `/stats/time-machine` | Get the collection as of a past date |
| `GET` | `/stats/time-machine/bounds` | Get the time machine's date range |

## Deep Identification

| Method | Path | Summary |
|---|---|---|
| `GET` | `/deep-identification/capability` | Report whether Deep Analysis is enabled for the current user |
| `GET` | `/deep-identification/jobs` | List the owner's Deep Analysis jobs |
| `POST` | `/deep-identification/jobs` | Start a Deep Analysis job |
| `GET` | `/deep-identification/jobs/{id}` | Get a Deep Analysis job |
| `DELETE` | `/deep-identification/jobs/{id}` | Delete a Deep Analysis job |
| `POST` | `/deep-identification/jobs/{id}/apply` | Confirm a Deep Analysis proposal through the existing write path |
| `POST` | `/deep-identification/jobs/{id}/cancel` | Cancel a Deep Analysis job |
| `GET` | `/deep-identification/jobs/{id}/events` | Stream a Deep Analysis job's events (SSE) |
| `PATCH` | `/deep-identification/jobs/{id}/proposal` | Save owner edits and per-field decisions on a Deep Analysis proposal |
| `POST` | `/deep-identification/jobs/{id}/retry` | Retry a Deep Analysis job |

## Agent

| Method | Path | Summary |
|---|---|---|
| `POST` | `/agent/chat` | Chat with coin search agent (streaming) |
| `GET` | `/agent/coin-search-prompt` | Get coin search prompt |
| `GET` | `/agent/coin-shows-prompt` | Get coin shows prompt |
| `POST` | `/agent/collection/proposals/{proposalId}/cancel` | Cancel collection proposal |
| `POST` | `/agent/collection/proposals/{proposalId}/commit` | Commit collection proposal |
| `GET` | `/agent/conversations` | List saved conversations |
| `POST` | `/agent/conversations` | Save conversation |
| `GET` | `/agent/conversations/{id}` | Get conversation |
| `DELETE` | `/agent/conversations/{id}` | Delete conversation |
| `GET` | `/agent/models` | List available Anthropic models |
| `GET` | `/agent/portfolio-summary` | Get portfolio summary for AI analysis |
| `GET` | `/agent/status` | Get agent status |
| `GET` | `/agent/valuation-prompt` | Get valuation prompt |
| `POST` | `/coins/{id}/estimate-value` | Queue coin value estimate |

## Coin Copilot

| Method | Path | Summary |
|---|---|---|
| `GET` | `/agent/copilot/capability` | Resolve Coin Copilot or legacy mode |
| `POST` | `/agent/copilot/runs` | Start a durable Coin Copilot run |
| `GET` | `/agent/copilot/runs/{runId}` | Read an owner-scoped Coin Copilot run |
| `POST` | `/agent/copilot/runs/{runId}/cancel` | Cancel a Coin Copilot run |
| `GET` | `/agent/copilot/runs/{runId}/events` | Replay and follow Coin Copilot events |
| `POST` | `/agent/copilot/runs/{runId}/resume` | Resume a paused Coin Copilot run |
| `GET` | `/agent/copilot/threads/{threadId}` | Read an owner-scoped Coin Copilot thread |

## Health (collection quality)

| Method | Path | Summary |
|---|---|---|
| `GET` | `/coins/health` | List coin health scores |
| `GET` | `/coins/{id}/health` | Get metadata health for a single coin |
| `GET` | `/stats/health` | Get collection health summary |

## Admin

| Method | Path | Summary |
|---|---|---|
| `GET` | `/admin/availability-cycles` | List availability cycles |
| `GET` | `/admin/availability-cycles/{id}` | Get availability cycle detail |
| `GET` | `/admin/availability-runs` | List legacy availability check runs |
| `GET` | `/admin/availability-runs/{id}` | Get legacy availability run detail |
| `POST` | `/admin/availability/run` | Trigger manual wishlist availability check |
| `GET` | `/admin/backups` | Full backups |
| `POST` | `/admin/backups` | Make a full backup now |
| `PUT` | `/admin/backups/upload` | Import a full backup |
| `GET` | `/admin/backups/{name}` | Download a full backup |
| `DELETE` | `/admin/backups/{name}` | Delete a full backup |
| `POST` | `/admin/backups/{name}/restore` | Restore a full backup |
| `POST` | `/admin/collection-currency` | Change the collection currency |
| `GET` | `/admin/deep-identification/observability` | Get Deep Identification observability |
| `GET` | `/admin/health/summary` | Get admin collection health summary |
| `GET` | `/admin/logs` | Get application logs |
| `GET` | `/admin/numista/health` | Get Numista operational health |
| `GET` | `/admin/numista-values` | Numista values batch status |
| `POST` | `/admin/numista-values/run` | Run a Numista values batch now |
| `GET` | `/admin/settings` | Get application settings |
| `PUT` | `/admin/settings` | Update application settings |
| `GET` | `/admin/settings/defaults` | Get setting defaults |
| `GET` | `/admin/test-anthropic` | Test Anthropic connection |
| `GET` | `/admin/test-searxng` | Test SearXNG connection |
| `GET` | `/admin/users` | List all users |
| `DELETE` | `/admin/users/{id}` | Delete a user |
| `POST` | `/admin/users/{id}/reset-password` | Reset user password |
| `PUT` | `/admin/users/{id}/role` | Update user role |
| `GET` | `/admin/valuation-runs` | List valuation runs |
| `POST` | `/admin/valuation-runs/trigger` | Trigger manual valuation |
| `GET` | `/admin/valuation-runs/{id}` | Get valuation run detail |
| `POST` | `/admin/valuation-runs/{id}/cancel` | Cancel a valuation run |
| `GET` | `/admin/wishlist-search-alert-runs` | List all wishlist search alert runs |

## Admin Security

| Method | Path | Summary |
|---|---|---|
| `GET` | `/admin/security/events` | List security events |
| `GET` | `/admin/security/exposure-check` | Get exposure check |
| `GET` | `/admin/security/summary` | Get security summary |
| `POST` | `/admin/users/{id}/unlock` | Unlock user |

## Internal

| Method | Path | Summary |
|---|---|---|
| `POST` | `/internal/tools/collection_summary` | Get collection aggregate summary |
| `POST` | `/internal/tools/commit_update` | Commit an update proposal |
| `POST` | `/internal/tools/get_coin` | Get a single coin by ID |
| `POST` | `/internal/tools/numista_detail` | Deep-identification Numista detail tool |
| `POST` | `/internal/tools/numista_search` | Deep-identification Numista search tool |
| `POST` | `/internal/tools/propose_update` | Create an update proposal |
| `POST` | `/internal/tools/search_my_collection` | Search user's collection |
| `POST` | `/internal/tools/top_coins_by_value` | Get top coins by current value |
