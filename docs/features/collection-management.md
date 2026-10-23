# Collection Management

> Browse, filter, group, search and organize your coins and banknotes.

## Overview

The collection has three views:

- **Collection overview** (`/collection`) — totals, the releases (each added to the collection with its Numista details, see [Numista Catalog Lookup](numista-integration.md#new-coins-from-the-releases)), the duplicates to trade (**Show in the gallery** opens the coin gallery, or the banknotes, with only them; **Share list** makes a file of them for other collectors, see [Sharing the Lists](wish-list.md#sharing-the-lists))
- **Coin gallery** (`/`) — every collection coin, with search, filters, grouping and select mode
- **Banknote gallery** (`/banknotes`) — the same for banknotes: **View by** **All banknotes** (one list, without groups, in the order chosen), or in groups by country (each with its flag, the first choice), monetary system, year or rarity, with how many in each; search, sort, country, **Banknote type** (as Numista classifies it, when it says) and grid or list. The list has no presentation column

## Coin Gallery

### Search and filters

The search takes the first row; below it, every field has the same size, in equal columns:

- **Face value** — the collection's face values, most common first, with how many coins of each
- **Sort** — recently updated, newest or oldest first, value, random
- **View by** — all coins, or grouped by country, period, monetary system, year, rarity or presentation. **Period** groups each country's coins by period, most recent first, as uCoin lists them: **Portugal · República Portuguesa 1986–2001**, **Portugal · União Europeia (Euro) 2002–2024**, **Reino Unido · Rainha Isabel II 1971–2008**, with the country's flag; a period whose name gives no years shows those of its coins, and coins without a period come last in their country (see [Period](coin-details.md#period)). A monetary system is named as Numista names its currency, with its dates (**Escudo (1911-2001)**, **Real (decimalizado, 1835-1910)**, so the réis of Portugal and the real of Brazil stay apart); every euro coin, whatever its country, is in **Euro**. Coins without one come last
- **Country** — each with its flag, including *Extinct countries* (USSR, Yugoslavia, Ottoman Empire…)
- **Period** — once a country is chosen, its periods, the most recent first, with how many coins each has (**República Portuguesa 1986–2001 (48)**, **União Europeia (Euro) 2002–2024 (55)**); choosing another country starts the choice again. `GET /api/coins/periods?country=` lists them; `GET /api/coins?period=` filters by one
- **Coin type** — standard circulation, circulating commemorative or non-circulating, as Numista classifies it
- **Presentation** — loose, coincard, capsule…
- **Rarity**
- **Set** — shown when you have standard sets
- **Layout** — grid or list
- **Face** — obverse or reverse on the cards

Search covers the name, denomination, ruler, era, mint, inscriptions, notes and country.

A filter turned on from the collection overview (current or commemorative coins, only the duplicates to trade) shows as a chip above the gallery; its × removes it.

The country's flag also shows in the titles of the view by country (coins, banknotes and set albums), and in the list layout of coins and banknotes alike. The flags are the system's emoji, so they need no download; Chrome and Edge on Windows show the two letters of the country code instead.

Countries that no longer exist have no emoji. Those that flew a living country's flag show its emoji everywhere: Czechoslovakia the Czech Republic's, the Ottoman Empire Turkey's, and colonies their mother country's (Portuguese India 🇵🇹, Dutch East Indies 🇳🇱, French West Africa 🇫🇷, British India 🇬🇧…). For the others (USSR, Yugoslavia, East Germany, Netherlands Antilles, French Indochina, Zaire, Rhodesia, South Vietnam, the German states and more) the app draws the flag, shown in the titles, the list and the galleries' **Country** menu. That menu is the app's own, not the browser's, so its options can show the drawings: it opens with a click or the arrow keys, moves with the arrows, Home and End, jumps to a typed letter, chooses with Enter and closes with Escape.

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
