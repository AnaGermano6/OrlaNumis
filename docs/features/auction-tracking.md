# Auction Tracking (removed)

> **Status (2026-10):** removed. The Auctions page and the NumisBids/CNG account
> integration (credentials, watchlist sync, new-lot/bidding notifications and the
> Pushover watch-bid digest) went first; the rest followed: auction lots, price
> alerts, bid reminders, ending-soon alerts, the [auction calendar](auction-calendar.md),
> the `list_auction_lots`, `get_auction_lot` and `auction_counts` MCP tools, and
> their admin schedules and settings.

The `/api/auctions`, `/api/alerts`, `/api/reminders`, `/api/calendar` and
`/api/admin/auction-*` endpoints answer 404, and the `/auctions` page opens the
gallery.

## Saved data

Lots, alerts, reminders and calendar events saved before stay in their tables
(`auction_lots`, `price_alerts`, `bid_reminders`, `auction_events`,
`auction_alert_runs`, `auction_ending_runs`), which are no longer migrated.
Deleting a coin clears the old lots' link to it, and deleting a user removes
their old rows.

## What stays

The AI agent's auction search is separate: the auction search sources in
**Admin → System**, [Similar Lots](similar-lots.md) and
[Price Trends](price-trends.md) keep working.

See also: [Wish List](wish-list.md), [Notifications](notifications.md)
