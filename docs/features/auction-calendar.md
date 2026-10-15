# Auction Calendar (removed)

> **Status (2026-10):** removed with [auction tracking](auction-tracking.md). The
> calendar page, its custom events, the chat's **Save to Calendar** button on
> coin shows and the calendar pins in Quick Access are gone; `/api/calendar`
> answers 404 and `/calendar` opens the gallery.

The API drops the old `auction_events` table, with the events saved in it, at
startup.

See also: [Notifications](notifications.md)
