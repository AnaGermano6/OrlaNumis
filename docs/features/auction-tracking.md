# Auction Tracking

> **Status (2026-10):** the Auctions page and the NumisBids/CNG account integration (credentials, watchlist sync, new-lot/bidding notifications and the Pushover watch-bid digest) were removed. Lots saved before stay in the database and on the calendar; the sections below describe what remains.

## Event-link results

Single-lot linking validates ownership of both the lot and calendar event and commits only after its response can be reloaded. Bulk linking uses **partial success**: each distinct lot commits independently. The response contains `updated` (confirmed writes only) and `failures` with `lotId`, `code` (`not_found` or `storage_error`), and a safe error message. Successful links are kept; failed lots remain selected in the UI for retry. A null event ID unlinks the owned lot.

> Monitor NumisBids and CNG Auctions lots with provider-aware status tracking, price alerts, bid reminders, calendar links, and collection conversion.

## Overview

Track auction lots from NumisBids and CNG Auctions with status updates, price monitoring, and conversion to your collection when won. The two providers do not expose the same data: CNG Auctions supports richer hosted-auction sync and outcome automation where the provider reports the necessary signals, while NumisBids currently supports watchlist/import tracking only.

## Provider Capability Model

| Capability | CNG Auctions | NumisBids |
|---|---|---|
| Manual lot import from a lot URL | Supported | Supported |
| Current bid and hosted bid metadata | Synced where CNG exposes it | Best-effort listing data only |
| Max-bid tracking | Synced where CNG exposes absentee bid data | Manual entry required unless future NumisBids data exposes it |
| Won/lost/final outcome automation | Auto-detected where CNG reports closed-lot winner data | Manual status update required |
| Needs-attention flag after close | Supported as a reminder to verify unresolved lots | Supported and expected for unresolved lots |

NumisBids lots should be treated as tracked watchlist/import records. After the sale closes, check the provider page and update the lot status manually to **Won**, **Lost**, or **Passed**. If you won the lot, enter the winning bid before converting it to a collection coin.

Manual and scheduled sync use the same provider normalization and persistence.
CNG bid-bearing lots stay unresolved until the provider reports closure and both
the account and winning bidder identities are known. A past end date alone cannot
establish a win or loss; a later successful sync can resolve the outcome.
Previously recorded Won/Lost outcomes remain protected from automatic changes.
Manual sync still sends no notifications and only fetches the selected provider.
Sync failures are reported rather than presented as complete success; lots already
saved before a failure remain saved and can safely be refreshed on retry.

## Key Features

- **Manual Lot Entry** — Paste NumisBids or CNG lot URLs to add lots
- **Outbid Alerts** — A batched alert when CNG reports someone else holding the winning bid on a lot you are bidding on, once per time you lose the lead
- **Status Workflow** — Watching → Bidding → Won/Lost/Passed, with CNG outcomes auto-detected where available and NumisBids outcomes updated manually
- **Price Alerts** — Notify when bidding crosses your threshold
- **Bid Reminders** — Get reminded X minutes before lot closes
- **Won → Collection** — Convert lots marked **Won** into collection coins
- **AI Auction Search** — Ask the agent to find similar lots
- **Filtered Views** — Filter by status and source with badge counts
- **Lot Calendar** — Visual calendar of lot end dates

## Configuration

### Provider Sources
- Existing NumisBids lots use `source = numisbids`
- CNG lots use `source = cng`
- Lot URLs are stored in provider-aware source URL fields while preserving legacy NumisBids URL compatibility
- `statusSource` records whether a terminal won/lost status came from sync (`sync`) or an explicit user override (`manual`)

### Auction Calendar
- View auction lots on monthly calendar
- Add custom events with title, date, optional URL
- Filter by date range

### Price Alerts
- Set target price and direction (above/below)
- Auto-notify when threshold crossed
- The notification names the lot and its linked lot number, the sale, your target, and the current high bid against that target
- Triggered status prevents duplicate notifications

### Bid Reminders
- Configurable lead time (e.g., 15 minutes before close)
- Notifications appear in in-app inbox

## API Endpoints

Full list and all details are in the main [features.md](../features.md#auction-tracking).

See also: [Wish List](wish-list.md), [Collection Management](collection-management.md)
