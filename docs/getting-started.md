# Getting Started Guide

This guide walks you through setting up Orla Numis for the first time, adding your first coin, and importing your collection.

## First-Time Setup

### Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `JWT_SECRET` | **Yes** | — | JWT signing key, 32+ characters. In release mode (`GIN_MODE=release`) the app stops on startup without it. |
| `AGENT_INTERNAL_SERVICE_TOKEN` | **Yes** (with the AI agent) | — | Shared API ↔ agent credential. Set the same high-entropy value for the Go API and the Python agent. |
| `CORS_ORIGINS` | No | localhost origins outside release mode | Comma-separated browser origins allowed to call the API |
| `TRUSTED_PROXIES` | Release mode | — | Your reverse proxy's IP/CIDR, or `none` when there is no proxy |
| `AGENT_SERVICE_URL` | No | `http://localhost:8081` | URL of the Python agent service |
| `AUREARIA_BACKUP_DIR` | No | `~/Aurearia-backups/<compose project>` | Where the full backups are kept (Docker Compose) |

`task init` creates `.env` with a random `JWT_SECRET` and `AGENT_INTERNAL_SERVICE_TOKEN`. Run `task --list` to see every Taskfile command.

### 1. Start the Application

With Docker Compose (the app and the AI agent, built from this checkout):

```sh
task init
docker compose -f docker-compose.dev.yml up -d --build
```

The app is at `http://localhost:8080`. Data and uploads live in Docker volumes; full backups go to `~/Aurearia-backups/aurearia-dev`.

For development without Docker:

```sh
task up        # Go API on :8080 and the Vite dev server on :5173
task up-all    # the same plus the Python agent on :8081
```

### 2. Create Your Admin Account

Open the app and click **Register**. The **first user** becomes the **admin**. After that, registration is closed unless an administrator sets `RegistrationMode` to `open`.

On the first sign-in a prompt offers the **Getting Started** guide; it is always in **Settings → Help**.

### 3. Configure Settings (Admin)

In **Admin**:

- **AI** — choose **Anthropic** (recommended) or **Ollama**, set the API key or URL and the model, and optionally customize the prompts. The AI features stay off until a provider is configured.
- **System** — the public app URL, the log level, the search sources, and the **Numista API key** for catalog lookups (free at [numista.com/api](https://en.numista.com/api/), 2,000 requests a month), plus the collection currency.
- **Coin Properties** — the categories (`Modern` and `Other` by default) and eras offered in the coin forms.
- **Schedules** — the automatic full backup (on by default, every day at 03:00 server time) and the wish list checks.

See [Admin Settings](features/admin-settings.md).

### 4. Start Adding Coins

Click **Add Coin**:

- **Catalog** — search Numista and fill the coin from its catalog entry; add your price and date, and save.
- **Manual Mode** — fill in the fields yourself; only **Name** is required.
- **AI Assist Mode** — photograph the obverse (required), the reverse and a coin card, and review the draft the AI prepares before saving. Needs an AI provider.

Choose **Add to Wishlist** if you don't own the coin yet. After saving, upload or replace images from the coin's page.

Banknotes have their own form and gallery (**Collection → Banknotes**).

---

## Adding Coins

### Assisted Capture

**AI Assist Mode** uses the same step-by-step layout as **Identify Coin**: **Obverse**, **Reverse**, then **Card**. Only the obverse is required; generate an intake draft immediately or continue to the optional reverse and coin-card steps.

In the installed app, the camera starts only when you tap **Start Camera**; library uploads remain available if access is denied. **Use manual mode instead** is always available, including for PDF coin cards.

**Generate Intake Draft** prepares editable suggestions; it does not add a coin to your collection. Review the draft, then explicitly **Confirm and Save Coin**.

### Fields

| Field | Description | Example |
| ----- | ----------- | ------- |
| Name | **Required.** A descriptive name | `2 Euro 2007 Treaty of Rome` |
| Category | One of the configured categories | `Modern` |
| Material | Gold, Silver, Bronze, Copper, Electrum, Other | `Other` |
| Denomination | Face value | `2 Euro` |
| Ruler / Issuer | Issuing authority | `Portugal` |
| Era | One of the configured eras | `modern` |
| Year | Year of issue | `2007` |
| Country | Issuing country | `Portugal` |
| Mint | Mint location | `Lisbon` |
| Weight (grams) | Coin weight | `8.5` |
| Diameter (mm) | Coin diameter | `25.75` |
| Grade | Condition grade | `UNC`, `VF` |
| Presentation | Loose, coincard, capsule… | `coincard` |
| Quantity | How many you hold | `1` |
| Obverse / Reverse Inscription | Legends | |
| Obverse / Reverse Description | Designs | |
| Purchase Price, Current Value | In the collection currency | `3`, `4` |
| Purchase Date | Date acquired | `2024-03-15` |
| Store | Dealer or seller | `Coin shop` |
| Notes | Free-text notes (Markdown supported) | |

### Uploading Images

Saving the record and uploading its images are separate steps. If an upload fails after the record is saved, **Retry remaining uploads** resumes only the unfinished uploads without creating another coin; **Open saved coin** lets you finish from the coin's page instead.

AI intake era suggestions must match a configured era. An unsupported suggestion is kept in Notes and the review shows a warning; choose a configured era or leave it **Unknown** before confirming.

Coins without a photo can get their official images (ECB, INCM, Monnaie de Paris, Münze Österreich) from the coin's page.

---

## Import & Export

Import and export live in **Settings → Backups**. See [Data Import/Export](features/import-export.md).

### Exporting Your Collection

**Export ZIP** downloads your collection's data and photos in one archive. **Export PDF** makes a catalog for insurance or provenance.

### CSV Import (recommended for a first import)

Download the **CSV template** from **Settings → Backups**, fill it in a spreadsheet, export it as CSV and import it. Only `name` is required.

| Column | Example |
| --- | --- |
| `name` | `2 Euro 2007 EU Presidency` |
| `category` | `Modern` |
| `material` | `Other` |
| `denomination` | `2 Euro` |
| `ruler` | `Portugal` |
| `era` | `modern` |
| `mint` | `Lisbon` |
| `weightGrams` | `8.5` |
| `diameterMm` | `25.75` |
| `grade` | `UNC` |
| `purchasePrice` | `3` |
| `currentValue` | `4` |
| `purchaseDate` | `2024-03-15` |
| `purchaseLocation` | `Coin shop` |
| `isWishlist` | `false` |

Boolean columns accept `true/false`, `yes/no` or `1/0`. Dates use `YYYY-MM-DD`. Every column and its accepted aliases are listed in **Settings → Help**.

```csv
name,category,material,denomination,ruler,era,mint,weightGrams,diameterMm,grade,purchasePrice,currentValue,purchaseDate,purchaseLocation,isWishlist
2 Euro 2007 EU Presidency,Modern,Other,2 Euro,Portugal,modern,Lisbon,8.5,25.75,UNC,3,4,2024-03-15,Coin shop,false
1 Escudo 1986,Modern,Other,1 Escudo,Portugal,modern,Lisbon,1.7,16,VF,0.5,1,2025-01-20,Local show,false
```

### JSON Import

The import also accepts a **JSON array of coin objects**, with the same field names as above. At minimum each coin needs a `name`.

```json
[
  { "name": "1 Escudo 1986" },
  { "name": "2 Euro 2007 EU Presidency", "category": "Modern", "denomination": "2 Euro", "year": 2007, "isWishlist": false }
]
```

- `id`, `userId`, `createdAt` and `updatedAt` are ignored: new IDs are assigned and the coins belong to your account.
- Images are not imported; upload them afterwards.
- Each coin is imported on its own: one failing does not stop the others. The response says how many were imported.
- Duplicates are not detected: importing the same file twice creates the coins twice.

Through the API:

```sh
curl -X POST http://localhost:8080/api/user/import \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d @my-coins.json
```

### uCoin

Collections exported from uCoin.net can be added with **Settings → Backups → Import from uCoin** and completed from Numista.

---

## AI Coin Analysis

With an AI provider configured, a coin's page offers **Analyze with AI** for its obverse and reverse photos, and a grade estimate. The request goes through the Python agent to Anthropic or Ollama. See [AI Coin Analysis](features/ai-analysis.md).

To use Ollama:

```sh
ollama pull llava
ollama serve
```

Then set the Ollama URL in **Admin → AI** (default `http://localhost:11434`).

---

## Pre-commit hooks

Install the local hook runner once:

```sh
pip install pre-commit
pre-commit install
```

This enables the hooks in [`.pre-commit-config.yaml`](../.pre-commit-config.yaml): secret scanning and quick language-specific checks before each commit.
