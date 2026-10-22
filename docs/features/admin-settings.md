# Admin Settings

> Configure AI providers, manage users, set up integrations and schedules, and watch security and health.

## Overview

Administration is available to admins only. The first registered user is assigned admin automatically. The tabs are grouped into **Configuration** (Users, AI, System, Coin Properties, Catalogs) and **Operations** (Security, Schedules, Health, Logs).

## Users

- View all registered users, their role and creation date
- Edit a user, change their role, reset their password
- Unlock a user locked out after failed sign-ins
- Delete a user account (with confirmation)

## AI Configuration

**Provider Selection:**
- Choose **Anthropic Claude** (recommended) or **Ollama**
- AI features stay unavailable until a provider is configured

**Anthropic Setup:**
1. Get an API key: [console.anthropic.com](https://console.anthropic.com/)
2. Paste it into **Anthropic API Key**
3. Pick the model (the list comes from the API)
4. (Optional) Customize the agent prompts. The built-in ones are written for this collection (euro coins, escudos and réis, world coins and banknotes) and answer in European Portuguese; **Revert to Default** brings one back

**Ollama Setup:**
1. Run Ollama: `ollama serve`
2. Set **Ollama URL** (default: http://localhost:11434)
3. Set the **Vision Model** (default: llava)
4. Set the timeout (default: 300s)
5. Set the **SearXNG URL** for web search (default: http://localhost:8888)

## System

- **Public App URL** — the full browser URL of the app, used in outbound links
- **Log Level** — trace, debug, info, warn, error
- **Search Sources** — the dealer and auction hostnames the coin search, wishlist alerts and auction search may use
- **Numista** — the API key, lookup limits, operational health, and the requests made this month against the monthly limit (`NumistaMonthlyLimit`, 2,000 by default). From 90% of the limit, background requests (catalogue searches, market values, country filling) wait for the next month; your own lookups go on.
- **Coin Copilot (beta)** — enable the read-only multi-step assistant, its limits, and the hand-off to Deep Analysis
- **Deep Analysis** — enable the background identification workflow and set its limits
- **Collection currency** — the currency every value is kept and shown in

## Coin Properties

- **Categories** — one category per line, shown in the coin forms. The defaults are `Modern` and `Other`.
- **Eras** — one era per line. The defaults are `ancient`, `medieval` and `modern`; the coin form adds *Unspecified*.
- **Portuguese names** — a name per option, shown instead of the value when the site is in Portuguese
- **Custom Locations** — global mint coordinates and aliases used by the mint map

## Catalogs

Manage the catalog definitions used by structured references:

- Catalog labels and metadata
- Catalogs that require a volume (none of the built-in ones, which are for modern coins and banknotes)
- The same rules apply to create and edit validation

## Security

- **Security Overview** — failed sign-ins, locked accounts and recent events
- **Public-Facing Readiness** — checks to run before exposing the app outside a home network (trusted proxies, CORS, public URL, registration, internal tokens)
- **Security Events** — filter by type, user, IP and date
- **Locked Users** — unlock accounts
- **Registration** — shown in the readiness checks. The `RegistrationMode` setting is `closed` by default: after the first user, new accounts are refused unless it is `open`.

Repeated password failures from one IP block that IP for a while automatically.

## Schedules

Background jobs, with their start time read in the **Schedule Time Zone**:

- **Wishlist Availability Check** — checks the wish list's listing URLs, with run history and a manual run
- **Wishlist Search Alerts** — scheduled discovery for saved search alerts, with run history
- **Collection Valuation** — scheduled valuation runs, manual trigger and cancellation (needs AI)
- **Automatic Full Backup** — the daily backup of the whole database and every upload: on or off, start time, and how many automatic backups to keep (7 by default). See [Data Import/Export](import-export.md).

## Health

Aggregate collection health across all users: median score, low-score coins and the most common missing fields.

## Logs

- Application logs as they occur
- Filter by level
- Auto-refresh
- Export to a text file

## API Endpoints

```
GET    /admin/users                       # List users
PUT    /admin/users/:id/role              # Update a user's role
POST   /admin/users/:id/reset-password    # Reset a password
POST   /admin/users/:id/unlock            # Unlock a user
DELETE /admin/users/:id                   # Delete a user

GET    /admin/settings                    # All settings
GET    /admin/settings/defaults           # Default settings
PUT    /admin/settings                    # Update settings
POST   /admin/collection-currency         # Change the collection currency

GET    /admin/logs                        # Application logs
GET    /admin/test-anthropic              # Test the Anthropic connection
GET    /admin/test-searxng                # Test the SearXNG connection
GET    /admin/numista/health              # Numista operational health

GET    /admin/security/summary            # Security overview
GET    /admin/security/events             # Security events
GET    /admin/security/exposure-check     # Public-facing readiness

GET    /admin/availability-cycles         # Availability check history
POST   /admin/availability/run            # Run an availability check now
GET    /admin/wishlist-search-alert-runs  # Search alert run history
GET    /admin/valuation-runs              # Valuation run history
POST   /admin/valuation-runs/trigger      # Run a valuation now
POST   /admin/valuation-runs/:id/cancel   # Cancel a valuation run

GET    /admin/backups                     # Full backups
POST   /admin/backups                     # Make a full backup now
PUT    /admin/backups/upload              # Import a full backup
GET    /admin/backups/:name               # Download a full backup
DELETE /admin/backups/:name               # Delete a full backup
POST   /admin/backups/:name/restore       # Restore a full backup

GET    /admin/health/summary              # Collection health summary
```

The complete list is in [`docs/openapi.json`](../openapi.json).

## Security Notes

- The first registered user becomes an admin automatically
- Admins can promote and demote other users
- Secrets such as provider keys are write-only once saved
- Most settings apply to the whole instance; personal preferences stay in each user's settings

See also: [AI Coin Analysis](ai-analysis.md), [Coin Lookup](coin-lookup.md), [Authentication](../authentication.md)
