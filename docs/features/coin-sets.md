# Coin Sets

> Group coins into sets: lists from the catalogue, ready-made challenges, rules, free sets, goals with a wish list, and AI proposals.

## Overview

**Sets** (`/sets`) lists your sets, each with its kind, coin count and, for lists to complete, how much is done. **Create set** offers four paths and two other kinds.

## Ways to create a set

### From the catalogue (checklist)
Open the **Commemorative catalogue** (`/sets/commemoratives`), choose a country (and a face value, if you like) and press **Create set**. The set lists every coin of the catalogue, marks the ones in your collection and shows the ones missing. It keeps itself up to date: coins added to your collection are marked, and new coins join the list when they come out.

### Challenges (checklist)
Ready-made goals:
- **A coin of each euro country** — one place per country, filled by the first euro coin you have
- **Joint issue** — the same €2 commemorative in every country that issued it (for example the 50 years of the Treaty of Rome), one place per country
- **A year's series** — a country's eight coins of one year, from 1 cent to €2
- **Coins of my year of birth** — every coin of your collection struck that year, now and later

### By filters (smart set)
Country, face value, years, kind of coin (circulation, commemorative, collector) and presentation. The form counts the matching coins as you go. Coins you add later that match join by themselves.

### Choose the coins (standard set)
A free set: add the coins you want, one by one, from the set's page or from a coin's page. Standard sets are also the ones offered by the gallery's set filter and by **Add to set** in select mode.

### Goal with wishlist (goal set)
Goal sets take both collection and wishlist coins. Completion is collection items divided by collection plus wishlist items: 2 in the collection and 5 on the wish list is 2 / (2 + 5) = 28.6%.

### AI assistant proposal (agentic set)
Describe the set you want and the agent proposes a roster of up to 300 slots. Creating an agentic set submits a proposal request; no set is created until the proposal is approved. Needs an AI provider.

## The set page

- **Album** — the set's coins and, for lists to complete, the missing ones. Show all, only those you have, or only the missing; group by year or country. Each country shows its flag: on the places of a challenge, on the titles of the view by country, and in what a list from the catalogue or a year's series is of.
- **Members** — for standard and goal sets, add coins with a picker that searches the whole collection (50 results a page), reorder them, or take them out.
- **Edit** — name, description, colour, and the filters of a smart set.
- **Set information** (`/sets/:id/info`) — analytics, the acquisition rate per month, the value trend over time, and **Capture Snapshot** to record the set's value now.

Moving straight from one set to another resets dialogs and selections; requests still running for the previous set cannot overwrite the new one.

## The commemorative catalogue

Each euro country, with its flag, and its commemorative €2 coins and €5, €10 and €20 collector coins, with those you have. Sources: the ECB, the shops of the INCM, Monnaie de Paris and Münze Österreich, and Numista.

- Filter by all, in my collection, or missing
- **Catalogue options** (the "…" on a coin): **Hide from the catalogue**, or **It is the same coin as…** when the catalogue lists one coin twice. The other coin stays in its place and counts as yours if you have either. The choices apply to the country's list, the counts of every country and the checklist sets made from the catalogue.
- Hidden and joined coins are listed below the list, each with a **Show** button.

## Agentic proposals

- The Go API queues the request; the Python agent drafts the scope, the slots and verification notes.
- A proposal is reviewed on its own page (`/sets/proposals/:id`): edit slots, regenerate with feedback, reject, or approve.
- Only approval creates the set. Collection coins are then matched to the slots automatically.

## Security

Every set endpoint requires authentication and only reaches the signed-in user's own sets.

## API Endpoints

```
GET    /api/sets                     # List sets
POST   /api/sets                     # Create a set
POST   /api/sets/import-csv          # Create a set from CSV
POST   /api/sets/preview-smart       # Count the coins matching smart-set filters
GET    /api/sets/:id                 # Set details
PUT    /api/sets/:id                 # Update name, description, colour, filters
DELETE /api/sets/:id                 # Delete a set

GET    /api/sets/:id/coins           # Member coins
POST   /api/sets/:id/coins           # Add a coin
PUT    /api/sets/:id/coins/order     # Reorder coins
DELETE /api/sets/:id/coins/:coinId   # Take a coin out
GET    /api/sets/:id/checklist       # A checklist set's slots
GET    /api/sets/:id/completion      # Completion details
GET    /api/sets/:id/analytics       # Analytics
GET    /api/sets/:id/trends          # Value trend
POST   /api/sets/:id/snapshot        # Capture a snapshot

GET    /api/coin-catalogue                    # A country's catalogue
GET    /api/coin-catalogue/countries          # The catalogue's countries
GET    /api/coin-catalogue/joint-issues       # Joint issues of the euro area
PUT    /api/coin-catalogue/overrides          # Hide a coin, or join it to another
DELETE /api/coin-catalogue/overrides          # Show a coin again

POST   /api/set-builder/runs                       # Submit an agentic proposal request
GET    /api/set-builder/proposals/:id              # Review a proposal
PUT    /api/set-builder/proposals/:id              # Edit a proposal
POST   /api/set-builder/proposals/:id/approve      # Approve and create the set
POST   /api/set-builder/proposals/:id/reject       # Reject
POST   /api/set-builder/proposals/:id/regenerate   # Regenerate from feedback
```

## Related Features

- [Collection Management](collection-management.md) — The set filter in the gallery
- [Bulk Operations](bulk-operations.md) — Add several coins to a set
- [Collection Statistics](statistics.md) — Aggregate analytics
