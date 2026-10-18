# Collection Management

> Browse, filter, group, search and organize your coins and banknotes.

## Overview

The collection has three views:

- **Collection overview** (`/collection`) — totals, the releases (each added to the collection with its Numista details, see [Numista Catalog Lookup](numista-integration.md#new-coins-from-the-releases)), the duplicates to trade (**Show in the gallery** opens the coin gallery, or the banknotes, with only them), and the coins still without a photo, with a button to fetch their official photos in batches
- **Coin gallery** (`/`) — every collection coin, with search, filters, grouping and select mode
- **Banknote gallery** (`/banknotes`) — the same for banknotes, in groups: **View by** country (each with its flag), year or rarity, with how many in each; search, sort, country, **Banknote type** (as Numista classifies it, when it says) and grid or list. The list has no presentation column

## Coin Gallery

### Search and filters

The search takes the first row; below it, every field has the same size, in equal columns:

- **Face value** — the collection's face values, most common first, with how many coins of each
- **Sort** — recently updated, newest or oldest first, value, random
- **View by** — all coins, or grouped by country, year, rarity or presentation
- **Country** — each with its flag, including *Extinct countries* (USSR, Yugoslavia, Ottoman Empire…), which have none
- **Coin type** — standard circulation, circulating commemorative or non-circulating, as Numista classifies it
- **Presentation** — loose, coincard, capsule…
- **Rarity**
- **Set** — shown when you have standard sets
- **Layout** — grid or list
- **Face** — obverse or reverse on the cards

Search covers the name, denomination, ruler, era, mint, inscriptions, notes and country.

A filter turned on from the collection overview (current or commemorative coins, only the duplicates to trade) shows as a chip above the gallery; its × removes it.

The country's flag also shows in the titles of the view by country, and in the list layout of coins and banknotes alike. The flags are the system's emoji, so they need no download; Chrome and Edge on Windows show the two letters of the country code instead.

Grouping by country fills in, once per browser session, the country of coins linked to Numista but saved without one.

### Select mode

Choose **Select coins** in the title bar (or **Select** in the installed app), pick coins, then:

- **Add to set** — adds the selected coins to one of your standard sets
- **Delete** — deletes them, after confirmation

The banknotes have the same: **Select banknotes** on their page, then **Select all** or one by one, **Add to set** or **Delete**.

### Random sort

Random sort uses a seed kept in `sessionStorage` (`coins:randomSeed`), so pagination stays stable within a session. Choosing Random again reshuffles; in the installed app, coming back after a while reshuffles too.

### Mobile and installed app

- Swipe gallery for browsing on phones, with the position kept across navigation
- Pull to refresh
- The filters open from a menu, full width

## How to Access

1. After signing in, the coin gallery opens
2. Use the search and the filters above the gallery
3. Click a coin card to open its details

## API Reference

```
GET    /api/coins                 # List coins (filters below)
POST   /api/coins                 # Create a coin
GET    /api/coins/:id             # Coin details
PUT    /api/coins/:id             # Update a coin
DELETE /api/coins/:id             # Delete a coin
POST   /api/coins/bulk            # Bulk actions: set, delete
GET    /api/coins/denominations   # Face values with their counts
GET    /api/coins/countries       # Countries with their counts
```

### Query parameters of `GET /api/coins`

- `kind` — `coin` or `banknote`
- `search` — full-text search
- `denomination`, `country`, `extinct`, `presentation`, `rarity`
- `circulation`, `commemorative`, `classification`
- `category`, `era`
- `set` — a set ID
- `repeated` — only coins held more than once
- `wishlist` — `true` for the wish list, `false` for the collection
- `sort`, `order`, `seed` — sorting
- `page`, `limit` — pagination (up to 100 per page)

The complete list is in [`docs/openapi.json`](../openapi.json).

## See Also

- [Coin Details](coin-details.md) — The full record of each coin
- [Coin Sets](coin-sets.md) — Organize coins into sets
- [Bulk Operations](bulk-operations.md) — Select mode
- [Collection Statistics](statistics.md) — Analyze your collection
