# Collection Statistics

> Charts, rankings and quality metrics of your collection.

## Overview

**Stats** opens a summary and expands into focused views. Wishlist coins and quick-capture drafts stay out of the collection figures.

| View | Route | Purpose |
|---|---|---|
| Summary | `/stats` | Summary cards, distributions, top coins and overview charts |
| Distribution | `/stats/distribution` | The collection by era and category, as a heat map |
| Timeline | `/stats/timeline` | When each coin joined the collection |
| Map | `/stats/mint-map` | Mint geography for coins with a mapped mint |
| Health | `/stats/health` | Collection health score and what to fix |
| Time Machine | `/stats/time-machine` | The collection as it stood on any past date — see [Time Machine](time-machine.md) |

## Summary

- **Summary cards** — coins owned, coins on the wish list, total value, total invested and ROI
- **Distributions** — by coin type, era, category, material and grade, plus the top rulers
- **Price ranges** — how many coins fall in each value band, in the collection currency
- **Top coins** — the most valuable coins
- **Acquisition flow** — coins bought by purchase period, ruler, era and type (only coins with a purchase date)

## Health

A score from 0 to 100 and a grade (A–F) for the collection's coins and
banknotes (not the wishlist), weighted across four dimensions:

- **Metadata** (40%) — face value, country, monetary system, category,
  material, grade and, for a coin, its mint (a banknote has a printer instead)
- **Image coverage** (20%) — two photos or more count in full, one by half
- **Valuation freshness** (20%) — a value checked within 30 days counts in
  full, less as it ages; a value of unknown age (no date of when it was set,
  nor a purchase date) counts as one over six months old. A purchase date is
  not needed.
- **AI coverage** (20%) — an AI analysis of each side

While no AI provider is set up (Admin → AI), no item can have an AI analysis:
that dimension is left out and the others share its weight (metadata 50%,
images 25%, valuation 25%), and no analysis is asked for. Missing items are
named in the app's language.

Below the score, a list of what needs attention, with quick actions (edit metadata, upload images, run a valuation, run an AI analysis). Administrators see the health of every collection in **Administration → Health**.

## Map

Mints with coordinates appear on the map with the coins struck there. Global mint locations are managed in **Administration → Coin Properties**; your own in **Settings → Data**.

## API Endpoints

```
GET    /api/stats                     # Summary statistics
GET    /api/stats/distribution        # Distribution by era and category
GET    /api/stats/health              # Collection health
GET    /api/stats/time-machine        # The collection as of a date
GET    /api/stats/time-machine/bounds # The time machine's date range
```

## Related Features

- [Collection Time Machine](time-machine.md)
- [Coin Sets](coin-sets.md) — Per-set analytics and value trend
- [Collection Management](collection-management.md)
