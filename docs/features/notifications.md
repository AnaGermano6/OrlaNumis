# Notifications

> In-app alerts for social interactions, price alerts, and reminders

## Overview

Open **Notifications** from the top bar or sidebar badge to review in-app notifications. The badge shows unread count and refreshes while signed in.

Polling starts after in-app login without a reload and runs once per minute with only one active timer. Logout, account changes, and app unmount clear the count and invalidate pending responses; an older overlapping request cannot replace a newer count. Recoverable count failures show a retry action while keeping the last count for the current account.

## Notification Types

- **Social activity** — Follow requests, accepted follows, comments, and ratings.
- **Wishlist availability** — Scheduled/manual availability checks can notify when wishlist items become unavailable.
- **Coin of the Day** — Daily featured coin notification for opted-in users.
- **Set milestones** — Coin set milestone notifications when configured thresholds are crossed.

## Actions

- Mark notifications as read
- Delete notifications
- Open linked detail views where a notification includes a reference
- Open Coin of the Day directly in the featured coin modal

## Related Features

- [Collection Management](collection-management.md)
- [Wish List](wish-list.md)
- [Coin Details](coin-details.md)

See also: [features.md](../features.md)
