# Data Import/Export

> Import coins from CSV, JSON or uCoin, export your collection, and back up the whole app.

## Overview

Import and export live under **Settings → Backups**; a spreadsheet of the coins is exported from the galleries.

## Spreadsheet export

**Export** in the coin gallery and in the banknote gallery (in the title bar on a computer, by the title in a phone's browser, under the filters in the installed app) downloads a spreadsheet:

- **What:** the gallery's coins or banknotes as its filters and search keep them (all of them when none is on), or **the whole collection**, coins and banknotes; in the gallery's order, the wishlist left out. Up to 10,000 items.
- **Excel (.xlsx)** — one sheet, its header bold and fixed with filters, numbers and dates as Excel's own (amounts with two decimals, dates as 21/10/2026 in Portuguese), so they sort and add up. It opens in Excel, Numbers, LibreOffice and Google Sheets.
- **CSV** — the same columns; in Portuguese with ";" between fields and a decimal comma, as Excel there expects, in English with "," and a point; dates as 2026-10-21.
- **Columns**, named in the app's language: ID, kind (coin or banknote), name, country, monetary system, face value, year, dates, mint, material, weight, diameter, grade, presentation, rarity, quantity, purchase price, date and place, current value and when it was updated (in the collection currency), references, Numista page, sets and notes; and, when there is a banknote, its series, serial number, signatures, printer, width and height. Materials, presentations and rarities are written in the app's language.
- A text that starts like a formula (`=`, `+`, `@`) is kept as text.

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
