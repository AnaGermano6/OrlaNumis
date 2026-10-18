# Bulk Operations

> Select several coins to add them to a set or delete them in one go.

## Overview

Use select mode in the coin gallery to apply the same action to several coins without opening each detail page.

## Key Features

- **Select mode** — Choose **Select coins** in the title bar (or **Select** in the installed app), then pick one or more coins. **Select all** and **Deselect all** work on the current page.
- **Banknotes** — **Select banknotes** in the Banknote gallery works the same way; **Select all** takes every banknote listed.
- **Add to set** — Adds the selected coins to one of your standard sets.
- **Delete** — Deletes the selected coins, after confirmation.
- **Installed-app friendly** — The action bar stays at the bottom of the screen, within reach.

The API behind it is `POST /api/coins/bulk`, with the actions `set`, `delete` and `export` (the selected coins as data).

## Related Features

- [Collection Management](collection-management.md)
- [Coin Sets](coin-sets.md)
- [Coin Details](coin-details.md)
