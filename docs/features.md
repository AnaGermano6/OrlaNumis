# Features

Orla Numis is a self-hosted catalogue for a personal coin and banknote collection. This page is the one-page overview; each area has its own page in [`features/`](features/INDEX.md).

The interface is in English and European Portuguese (pt-PT). Every value is kept and shown in one collection currency, chosen by the administrator.

## Collection

- **Coin gallery** — search, and filters of the same size in equal columns: face value, sort, view by (country, year, rarity, presentation), country (including extinct countries), coin type, presentation, rarity, set, grid or list, obverse or reverse. → [Collection Management](features/collection-management.md)
- **Banknotes** — their own form and gallery.
- **Collection overview** — totals, highlights, the releases and the duplicates to trade.
- **Coin details** — numismatic fields, quantity, images, catalog references, Numista market value, sets, activity journal and timeline. → [Coin Details](features/coin-details.md)
- **Select mode** — add several coins to a set, or delete them. → [Bulk Operations](features/bulk-operations.md)

## Adding coins

- **From the Numista catalog** — search Numista and fill the coin from its entry. → [Numista Catalog Lookup](features/numista-integration.md)
- **Manually**, or with **AI Assist Mode** from photos of the obverse, reverse and a coin card.
- **Identify Coin** — photograph a coin or an NGC slab at a show, verify the NGC certificate, and save to the wish list or collection. → [Coin Lookup](features/coin-lookup.md)
- **Quick Capture** — show-floor drafts with photos and notes, promoted later. → [Quick Capture](quick-capture.md)
- **Import** — CSV (with a template), JSON, or a uCoin.net export. → [Data Import/Export](features/import-export.md)

## Wish list

Coins you want to buy, a link to a dealer page turned into a wishlist coin, availability checks, saved search alerts with a review queue, and **Mark as purchased** to move a coin into the collection. → [Wish List](features/wish-list.md)

## Sets

Lists from the commemorative catalogue, ready-made challenges (a coin of each euro country, a joint issue, a year's series, coins of your year of birth), sets by filters, free sets, goals with a wish list, and AI proposals reviewed before they are created. The **Commemorative catalogue** lists each euro country's commemoratives and collector coins, with those you have; a coin can be hidden, or joined to another listed twice. → [Coin Sets](features/coin-sets.md)

## Stats

Summary cards, distributions, top coins, acquisition flow, distribution heat map, timeline, mint map, collection health with quick fixes, and the **Time Machine**, which shows the collection as it stood on any past date. → [Collection Statistics](features/statistics.md), [Time Machine](features/time-machine.md)

## AI (optional)

With Anthropic Claude or Ollama configured by an administrator:

- **Coin analysis** of obverse/reverse photos, and a **grade estimate**. → [AI Coin Analysis](features/ai-analysis.md), [AI Grading](features/ai-grading.md)
- **Deep Analysis** — a resumable background identification with Numista evidence, a cited report and fields you accept one by one. → [Deep Analysis](features/deep-analysis.md)
- **Coin Agent** — one chat for dealer and auction searches, coin shows, price trends, similar lots, gap analysis and questions about your collection. → [Coin Agent](features/ai-search-agent.md)
- **Coin Copilot (beta, off by default)** — a read-only, multi-step version of the chat that shows its plan and sources. → [Coin Copilot](features/coin-copilot.md), [How it works](features/how-coin-copilot-works.md)
- **Valuation** — scheduled value estimates for the collection.

## Backups and export

- **Export ZIP** of your collection and photos, and an **insurance/provenance PDF**. → [PDF Export](features/pdf-export.md)
- **Full backup** (administrators) — the whole database and every photo in one zip, made every day by itself (the last 7 kept), with download, import and restore. → [Data Import/Export](features/import-export.md)

## Installed app and mobile

Installable PWA with offline reading, swipe gallery, pull to refresh, camera capture and voice dictation for notes. → [PWA Features](features/pwa-features.md), [PWA Guide](pwa-guide.md)

## Administration

Users, AI providers, Numista (with the requests used this month), search sources, coin properties, catalogs, security (readiness checks, events, locked users), schedules, collection health and logs. → [Admin Settings](features/admin-settings.md)

## Sign-in and security

Username and password with JWT access and refresh tokens, account lockout, automatic blocking of IPs with repeated failures, rate limits, closed registration after the first user, and private media served only to their owner. → [Authentication](authentication.md), [Security Principles](security-principles.md), [Deployment](deployment.md)
