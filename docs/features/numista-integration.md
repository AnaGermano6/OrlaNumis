# Numista Catalog Lookup

> Find and explicitly save Numista entries as structured catalog references.

## Overview

Numista lookup is part of catalog-reference management rather than a
standalone destination. Saved coins use **Catalog References** as the canonical
surface. Identify Coin and Quick Capture retain lookup where it provides
context while reviewing a draft attribution.

Results are decision support only. Orla Numis saves a Numista reference only
after the collector selects a candidate and confirms the relevant save action;
it never attaches every result or infers a selection from result order.

## Setup

1. Obtain a Numista API key from [numista.com/api](https://en.numista.com/api/).
2. Paste the key in **Admin → System → Numista API Key**.
3. Save the setting.

The configured allowance is shared by collectors using the Orla Numis instance.
Fresh equivalent searches and catalog details may be reused across saved-coin
and photo-assisted lookup. Cache content is disposable and never becomes a
collection record.

## Query and Ranking Contract

- Orla Numis generates concise Numista text queries from ruler/issuer (or a
  short structured title fallback), reverse inscription/type, and a reliable
  mint. It does not put denomination, date/range, material, catalog
  references, obverse text, slab/label text, notes, or descriptive prose in
  the provider `q`.
- Those excluded fields remain in the bounded evidence sent to Orla Numis's
  deterministic scorer. Omitting a field from `q` does not discard it from
  candidate ranking.
- Exact normalized mint aliases `SMN` and `SMNT` map to `Nicomedia`.
  Normalization is limited to Unicode compatibility/case folding and removal
  of spaces, dots, and hyphens. Embedded, longer, unknown, or prose-like
  mintmarks are not inferred.
- Saved-coin and Quick Capture draft screens request the proposal from the Go
  API. Non-NGC photo analysis reuses the same Go builder and returns a
  proposal without searching Numista. The browser displays and edits proposals
  but does not independently assemble provider query text.
- A query is `generated` only while it remains an exact, server-verifiable
  `numista-query-v2` proposal. Once the collector changes it, it is
  `user-edited` for that panel session even if the text is later restored.
  Input that was not initialized by a proposal is `manual`.
- Manual and user-edited queries are searched exactly as submitted and receive
  one application search attempt. Only an empty, verified generated primary
  query may receive one distinct relaxed attempt. The relaxed query keeps the
  subject and reliable mint while omitting reverse terms; every other status
  and source performs one application search.
- `effectiveQuery` reports the actual primary or relaxed query behind the
  returned outcome. `querySource`, `searchAttempt`, and
  `searchAttemptCount` disclose the source and whether one or two application
  searches occurred. The editable control continues to show the collector's
  submitted primary query, with relaxed use disclosed separately.
- Ranking is application-owned and deterministic (`numista-v1`). Candidate
  cards explain meaningful matches, conflicts, and unavailable evidence.
  Scores support comparison; they are not an attribution or authenticity
  opinion.
- Numista telemetry records safe source/attempt enums and digest-only
  correlation. It does not record query/evidence text, mint aliases, images,
  slab text, credentials, or raw provider payloads.

## Saved Coins: Catalog References

1. Open a saved coin.
2. Find **Catalog References** on the coin detail page.
3. Select **Search Numista** beside **Add Reference**.
4. Review or edit the proposed query, then submit it explicitly.
5. Review the candidates and select one.
6. Confirm **Add selected reference**.

If an untouched generated query returns no candidates, Orla Numis may try its
single relaxed query and identifies that effective query in the result. An
edited query is never rewritten or relaxed.

The lookup expands inline without replacing manual reference management. After
the selected reference is persisted and the refreshed reference list contains
it, the lookup collapses and the new Numista reference appears in the list.

The existing **Actions** page remains compatible during this placement
transition, but it does not contain the full lookup panel. It provides only a
compact link back to the saved coin's Catalog References section.

Selecting a result changes only local review state. A `CoinReference` is
created only after **Add selected reference** succeeds. Orla Numis reconstructs
the canonical URL as
`https://en.numista.com/catalogue/pieces<identifier>.html`; manual reference
validation and deduplication remain authoritative.

## Completing Records from Numista

**Admin → Coin properties → Complete records from Numista** fills in coin
records from the Numista types the app has already saved, without asking
Numista: it costs none of the monthly quota.

- **What it fills:**
  - the **monetary system**, from the type's currency, when none is set, on
    coins and banknotes: Numista's full name of the currency, with its dates
    ("Escudo (1911-2001)"), and **Euro** for every euro coin, whatever the
    country (see [Monetary System](coin-details.md#monetary-system));
  - on coins, the **material**, from Numista's composition, when the coin's
    is Other or empty;
  - the **weight** and the **diameter**, when missing;
  - **coin card** as the presentation, when none is set and Numista's
    comments say the type is issued in one.
- **Which coins:** the user's coins and banknotes (collection and wishlist)
  with a Numista catalogue reference whose type details are saved; a
  banknote gets only its monetary system. A coin
  linked to a type the app has not saved yet is left out and counted.
- **Preview first:** the section opens with what would change: how many
  coins, how many of each field, the materials to set, and every coin with
  its values and Numista's composition. **Save** asks for confirmation and
  saves all of it at once.
- **Corrections:** coins whose material is set but differs from Numista's
  composition (for example, silver plated copper recorded as Silver) are
  listed apart and change only when **Also correct these coins** is ticked.

How a composition becomes a material:

- Alloys first: nordic gold, aluminium bronze (also aluminium-nickel-bronze
  and bronzital), stainless steel (also acmonital), nickel brass (also nickel
  silver), copper-nickel, billon.
- A coin of two metals is **Bimetallic**, whatever they are.
- A plated or clad coin counts by its core (silver plated copper is Copper,
  copper plated zinc is Zinc), except steel: copper, nickel, brass or bronze
  plated steel is **Plated steel**.
- The alloy in brackets ("Cuproníquel (75% copper, 25% nickel)") is not read.
- Plain steel or iron fits no material and stays as it is.

The same rules set the material when a coin is added from the Numista
catalogue.

## Daily Values from Numista

**Admin → Schedules → Numista Values** keeps the coins' current values in
step with Numista's estimated prices, a few coins a day, so the month's quota
lasts. It is off until it is turned on there.

- **Which coins:** every collection item (coins and banknotes, not the
  wishlist) linked to a Numista type, with a grade Numista prices: UNC (also
  BNC, MS, FDC, Soberba…), AU, XF, VF, F, VG and G. Coins without a grade, or
  graded Proof, wait for one; the card counts them.
- **The price:** Numista's price, in the collection currency, for the coin's
  grade on the issue of its year (the type's first issue when the year does
  not pick one). A grade without a price takes the nearest lower grade's,
  else the nearest higher one's.
- **The value:** the price replaces the coin's current value. The value it
  replaces stays in the coin's value history (a coin with no history first
  gets an entry for it), and the new one is recorded as **Numista price**. A
  price equal to the value only marks the value as checked.
- **How often:** each coin once every 30 days, those never checked first,
  then the longest unchecked. Prices the app fetched in the last 7 days and
  issues fetched in the last 30 cost no request.
- **Within the quota:** each day's batch may make what is left of the
  background share of the month's quota (90%, see the usage in Admin →
  System), spread over the days left in the month, and never more than
  **Numista requests per day** (40 by default). It stops when that is spent
  and goes on the next day.
- **Warnings:** from 75% of the quota the card and, while the batch is on,
  the collection overview (for an admin) say the quota is running low and the
  batches get smaller; at the pause they say until when the values wait (the first day of
  next month). A batch also stops by itself when Numista says the quota is
  spent, has no key, or does not answer three times in a row.
- **Run now** starts a batch at once, on or off, within the same budget. The
  card lists the latest batches: what each updated, kept, found without a
  price, could not look up, and the requests it made of its budget.

## New Coins from the Releases

**Add to collection** on a new coin of the collection overview's releases
opens Add Coin with what the ECB or the mint says, and looks up the coin's
Numista type:

- a coin listed from Numista is its own type, and a mint's coin the catalogue
  has matched to Numista keeps that type, with no new request;
- otherwise Numista's coins of the country, year and face value are searched
  (once a day at most for each), and the coin is the one whose ECB image or
  mint page is the coin's, or the only one whose name is the coin's.

When the type is sure, its details complete the form: metal, weight,
diameter, mint, sides and their lettering, the Numista and other catalogue
references, and Numista's notes. The coin keeps its own name, face value,
country, year, circulation, rarity and source page, and a field the user has
already changed is left alone. **Not this coin** lists the other candidates
and can go back to the release's details alone. When none is sure, the
candidates are listed to choose from, with a link to search the Numista
catalog. Nothing is saved until the coin is.

## Identify Coin and Quick Capture

1. Open **Identify Coin** and capture or upload photos.
2. Select **Analyze Photos**.
3. Review the extracted coin details.
4. If no usable NGC certification is found, review or edit the contextual
   server-generated Numista query and submit it when ready. Analysis creates
   evidence and a proposal only; it does not eagerly search Numista.
5. If a usable NGC result is found, NGC remains primary. Select
   **Also search Numista** to reveal the editable lookup. Revealing it does not
   call Numista; a request occurs only after explicit search submission.
6. Optionally select one Numista candidate and choose **Save as Draft**.

The selected reference is retained with the Quick Capture draft. The draft
list shows `Numista #<identifier>` when a selection exists and no Numista chip
when it does not. Promoting the draft to the collection or wish list copies
exactly the retained selection; saving or promoting without a selection adds
no Numista reference.

Draft updates have three explicit behaviors:

- omit selection fields to preserve the retained reference;
- provide a canonical selected ID/URL pair to replace it;
- send the explicit clear control to remove it.

Validation failures and unrelated edits do not silently replace or clear the
selection. Repeated promotion returns the already-created coin and does not
duplicate the reference.

## Lookup Outcomes

Expected provider/configuration outcomes return a typed state rather than
being collapsed into “no results”:

| Status | Meaning and next action |
|---|---|
| `success` | One or more usable candidates; review and select explicitly. |
| `empty` | Configured lookup completed with no usable candidates; revise the query. |
| `unconfigured` | Numista is not configured or rejected the credential; admins receive configuration guidance, other collectors are told to contact an administrator. |
| `quota-limited` | Numista returned 429; wait, using `retryAfterSeconds` only when the provider supplied a positive value. |
| `timeout` | The provider deadline elapsed; the query and selection remain available for retry. |
| `unavailable` | Network, provider, malformed-response, or lookup-capability failure; retry later. |

Malformed requests still use HTTP 400, authentication/authorization use
401/403, and unexpected application failures use a generic HTTP 500. Caller
cancellation ends the request and is not misreported as a provider status.

For broad lookup outcomes, `effectiveQuery` is the exact query used for the
final result. A generated empty primary may therefore return
`searchAttempt: "relaxed"` and `searchAttemptCount: 2`; manual, user-edited,
successful, configuration, quota, timeout, unavailable, and cancelled paths
do not receive that application-level fallback. Existing transient HTTP retry
behavior remains separate transport behavior.

## Freshness and Coalescing

Broad search responses may include:

- `cache.hit=true` when a fresh persisted in-memory entry was reused;
- `cache.coalesced=true` when the request joined equivalent in-flight work;
- `createdAt`, `expiresAt`, and `ageSeconds` for the shared result.

`hit` and `coalesced` are mutually exclusive. Expired entries are removed and
never served as current. Successful empty searches are cacheable; provider
failures are not. Current configuration is checked before cache access, so
removing the API key cannot be concealed by old cached results.

## Progressive Detail Enrichment

Broad candidates render first. The browser then submits the complete broad
candidate set to `POST /api/numista/enrich`; the server reranks it, chooses the
leading bounded subset, and fetches at most five details by default with
concurrency two. Client ordering or duplicate IDs cannot increase that limit.

Candidate detail states are:

- `not_requested` — broad result only;
- `enriched` — fresh provider detail applied;
- `cached` — fresh detail cache applied;
- `failed` — detail failed, but the broad result remains selectable.

Enrichment returns the complete candidate set and never silently changes the
collector's explicit selection. Partial or total detail failure does not turn
a successful broad search into `empty`.

## Navigation and Compatibility

- Orla Numis adds no standalone Numista route, sidebar item, or top-level menu
  item.
- Existing coin detail, Actions, Identify Coin, and Quick Capture routes remain
  available.
- Manual structured-reference add, edit, delete, validation, and
  deduplication remain authoritative.
- The legacy Numista search endpoint remains available for compatibility.
- Numista image search, image upload, and perceptual matching are not part of
  this integration. Coin photos may supply bounded text evidence through the
  existing analysis workflow, but no image is sent to a Numista search API.

## API Endpoints

```text
POST /api/numista/query-proposal      # Local versioned proposal; no provider call
POST /api/numista/lookup             # Typed explicit lookup
POST /api/numista/enrich             # Bounded progressive detail enrichment
GET  /api/numista/search?q=...       # Deprecated compatibility adapter
GET  /api/coin-releases/numista-type # The Numista type of a new coin, and its candidates
GET  /api/numista/catalog/types/:id/issues # A type's issues: dates, mint marks, a banknote's signatories, Pick number and kind (cached, no prices)
GET  /api/coins/numista-fill          # What completing coin records from saved types would change
POST /api/coins/numista-fill          # Complete them (?corrections=true also corrects materials)
POST /api/coins/lookup               # Analyze photos; does not eagerly search Numista
GET  /api/admin/numista/health       # Admin-only redacted rolling health
GET  /api/admin/numista-values        # Admin-only: the daily values batch, the quota and the latest batches
POST /api/admin/numista-values/run    # Admin-only: run a batch now
```

The deprecated GET adapter keeps its legacy `{count,types}` response and maps
non-success/non-empty domain outcomes to HTTP 503. New clients use the typed
POST routes.

`POST /api/numista/query-proposal` is an authenticated local helper for the
implemented direct/photo evidence contract. It validates bounded
`NumistaEvidence` and returns only `query`, `querySource: "generated"`, and
`generationVersion: "numista-query-v2"`. It does not call Numista, rank
candidates, persist data, or record provider telemetry.

## Measurement and Limitations

The sanitized six-case live comparison observed on 2026-08-12 improved
expected-candidate top-three inclusion from **1/6 (16.7%)** with the landed
verbose builder to **3/6 (50%)** with the v2 primary query: **+33.3 percentage
points**, with no previously top-three verbose case lost. The deterministic
24-known-coin scoring benchmark remained **24/24** without exact-ID evidence.

This is evidence for a narrower query transformation, not proof that concise
text always identifies a coin. Valentinian II and Trajan still missed the
expected candidate in the bounded first ten live results, while Athens
required the relaxed query to reach rank 3. Provider catalog/search ordering
may change after the observation date. Future tuning should make the smallest
measured transformation supported by sanitized fixtures and should not add
speculative alias grammars, scoring changes, extra retries, or image-search
infrastructure.

## Related Features

- [Coin Details](coin-details.md) — Structured catalog references
- [Coin Lookup](coin-lookup.md) — Photo analysis and NGC-first review
- [Quick Capture](../quick-capture.md) — Draft retention and promotion
- [Admin Settings](admin-settings.md) — Configure the Numista API key
- [API Reference](../api-reference.md#numista) — Typed request/response contract

See also: [Numista.com](https://en.numista.com/)
