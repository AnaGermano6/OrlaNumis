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

A score from 0 to 100 and a grade (A–F), weighted across four dimensions:

- **Metadata** — the main fields filled in
- **Image coverage** — obverse and reverse photos
- **Valuation freshness** — how recent the current value is
- **AI coverage** — coins with an AI analysis

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
