# Data Import/Export

> Import coins from CSV, JSON or uCoin, export your collection, and back up the whole app.

## Overview

Import and export live under **Settings → Backups**.

## Your collection

- **Export ZIP** — your collection's data and photos in one archive.
- **Export PDF** — a catalog for insurance or provenance, with photos and structured data. See [PDF Export](pdf-export.md).
- **Import** — coins from a JSON backup or a CSV file. Rows without a `name` are skipped, and the import says how many.
- **CSV template** — download it, fill it in a spreadsheet (Excel, Google Sheets, Numbers) and import it. The columns and their aliases are listed in **Settings → Help**.
- **uCoin import** — add a collection exported from uCoin.net and complete it from Numista (`/import/ucoin`).

## Full backup (administrators)

The **Full backup** section copies the whole app: the database (every user's data and the settings) and every upload, in one zip with a `manifest.json`.

- **Automatic** — made every day at the time set in **Administration → Schedules** (03:00 server time by default); the last 7 automatic backups are kept. A run missed while the computer slept is made when the app runs again.
- **By hand** — **Back up now**.
- **Where** — in a folder outside the containers: `~/Aurearia-backups/<compose project>` with `docker-compose.dev.yml`, or the folder set in `AUREARIA_BACKUP_DIR`. So the backups survive a deleted Docker volume. The folder is shown in the section.
- **Download** a backup, **import** one downloaded before (a file that is not a backup of the app, or is damaged, is refused), or **delete** one.
- **Restore** — after a confirmation, the backup is checked and the data of now is saved first in a *pre-restore* backup, so you can go back. The app then restarts and applies the backup; the page opens again by itself. Users may have to sign in again.

Manual, imported and pre-restore backups stay until you delete them. A backup holds every user's data, password hashes included: keep it safe.

## Related Features

- [Collection Management](collection-management.md)
- [PDF Export](pdf-export.md)
- [Admin Settings](admin-settings.md)

See also: [Getting Started](../getting-started.md), [Deployment](../deployment.md)
