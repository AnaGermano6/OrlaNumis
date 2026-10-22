# Coin Details

> Store comprehensive numismatic data, images, references, activity logs, and notes for each coin in your collection.

## Overview

Each coin in your collection can hold rich, structured metadata that captures its numismatic significance, provenance, valuations, and personal notes. The detail page is the primary interface for viewing and editing all coin information.

## Previous and Next

A coin's or banknote's page goes to the previous or the next one of the list
it was opened from, without going back to it: the coin gallery, the banknote
gallery or the wishlist, in that list's order (its search, filters, sort and
groups; an ungrouped gallery, its page). Beside **Back to the gallery** (on a
phone, in a row below it) are **‹**, the coin's place (**12 of 454**) and
**›**; the left and right arrow keys do the same, except while typing or with
a photo open. Each step replaces the page, so the browser's back still goes to
the list. The order is kept for the browser tab, so a reload keeps it; a coin
opened from somewhere else (a set, the overview) has them only when it is in
the last list seen.

## Core Fields

### Identification
- **Name** — Coin description or catalog name
- **Denomination** — Value (denarius, aureus, sestertius, etc.)
- **Ruler** — Authority or monarch on the coin
- **Era** — Time period (ancient, medieval, modern)
- **Category** — one of the admin-configured categories (Modern and Other by default)
- **Monetary system** — see below

### Monetary System

The currency system the coin or banknote belongs to, as Numista names it, with
its dates: **Escudo (1911-2001)**, **Real (decimalizado, 1835-1910)** (the réis),
**Libra esterlina (decimalizada, 1971-hoje)**. Every euro coin is **Euro**,
whatever its country. It is filled from Numista when a coin is added from the
catalogue, and for coins already saved by **Complete records from Numista**
(see [Completing Records from Numista](numista-integration.md#completing-records-from-numista));
it can also be typed, with the collection's systems suggested.

It takes the place of the era and the category, which ancient coins need and
a modern collection does not:

- **Galleries** — **View by** monetary system, coins and banknotes.
- **Cards** — the system without its dates ("Escudo"), with the full name on
  hover, in place of the category; the category shows while there is no
  system.
- **Statistics** — **By monetary system**, the twelve most common and the
  rest added up. The charts by category and by era, and their heat map, show
  only when the collection has more than one category or era.

### Physical Characteristics
- **Material** — gold, silver, billon, electrum, copper, bronze, aluminium bronze, brass, nickel brass, nordic gold, copper-nickel, nickel, bimetallic, aluminium, zinc, stainless steel, plated steel, or other. Coins linked to Numista can get it from Numista's composition (see [Completing Records from Numista](numista-integration.md#completing-records-from-numista))
- **Weight** — In grams
- **Diameter** — In millimeters
- **Thickness** — Optional thickness measurement
- **Grade** — chosen from a list, each grade with what it means: UNC, AU, XF, VF, F, VG and G, and the collector finishes BNC and Proof (PRF). A grade written otherwise (an import, the AI's "MS-65") stays selectable

### Numismatic Details
- **Mint Mark** — Mint location indicator
- **Obverse Inscription** — Text on the front
- **Reverse Inscription** — Text on the back
- **Structured Catalog References** — Formal references (RIC, RPC, SNG, Numista, etc.) with catalog, volume, number, optional invoice number, and authority URI

### Banknote Details

A banknote has, instead of a coin's mint, weight and diameter:

- **Series or issue** ("Chapa 13"), the **date printed on the note**, the
  **Pick number** ("176a": the note's number in the Standard Catalog of World
  Paper Money, a letter for each variant) and the issuing bank
- this very note's **serial number**, **signatures**, **kind of note**
  (ordinary, replacement or specimen), grade and a grading company's grade and
  certificate
- the paper: material, size, printer and watermark

**Numista issue** — a note linked to a Numista type can say which of its
issues it is. The issues of a banknote are told apart by the date printed and
the signatories (the 20 escudos of 13 September 1978 has twelve, one for each
pair of signatures), so **Choose the issue** lists them as "13/09/1978 ·
P176a · Vilar / Reis", and choosing one fills in the signatures (full names),
the Pick number with its variant letter, the printed date and the kind of
note. The issues come from the server's cache of Numista, loaded only when
asked for. Adding a banknote from the catalogue does the same with the issue
picked there. The issue chosen is saved with the note, and its market value is
that issue's.

### Financial Data
- **Purchase Price** — Original acquisition cost
- **Purchase Date** — When acquired
- **Current Value** — Latest estimated value
- **Dealer/Source** — Where acquired
- **Cost Basis** — Purchase price + any restoration costs

### Provenance & History
- **Provenance Notes** — Where the coin came from and any prior ownership
- **Condition Notes** — Any cleaning, restoration, or damage
- **Activity Journal** — Timestamped entries (e.g., "sent to NGC", "displayed at show", "cleaned")
- **Free-Text Notes** — Any additional information

## Images

### Image Management
- **Upload Multiple Images** — Support for obverse, reverse, edge, detail, full views
- **Upload Methods**:
  - Direct file upload
  - Paste image URL (via server proxy fetch)
  - Camera capture in PWA mode
- **Image Gallery** — Lightbox viewer with zoom and navigation
- **Set Primary Image** — Choose which image appears on coin cards in collection view

### Image Operations
- **Background Removal** — Remove image backgrounds in-place using client-side ML
- **Circle Clipping** — Automatically clip obverse/reverse to circular framing (gated behind optional `circleClip` parameter)
- **OCR Text Extraction** — Extract text from store cards, certificates, or labels

## AI Analysis

### Analysis Workflow
1. Upload photos of obverse and reverse
2. Click **Analyze with AI**
3. Choose provider (Anthropic Claude or Ollama)
4. System analyzes each side separately with tailored prompts
5. Review analysis and accept/reject to save

### Analysis Content
Includes identification, description, inscriptions, condition assessment, historical context, and estimated market value. Stored separately for obverse and reverse.

### Configuration
- **Admin → AI Configuration** — Select provider and model
- **Custom Prompts** — Customize analysis prompts for obverse, reverse, and text extraction

## Structured Catalog References

### Reference Management
- **Add Multiple References** — Track coins across different catalogs (RIC, RPC, SNG, Numista, etc.)
- **Reference Fields**:
  - Catalog (required): which reference catalog
  - Volume (required for some catalogs)
  - Number: the catalog entry number
  - Invoice number and authority URI when available
  - Authority URI: optional link to the authoritative source

### Built-in Catalogs
- **RIC** — Roman Imperial Coinage (requires volume)
- **RPC** — Roman Provincial Coinage (requires volume)
- **SNG** — Sylloge Nummorum Greacorum (requires volume)
- **Numista** — Numista catalog ID
- **Other** — Custom catalog references

### Numista Integration
- Direct search from the coin detail page
- Find matching coins by name, denomination, and ruler
- View thumbnails and link to full Numista entries
- Auto-suggest catalog references

## Activity Journal

Track the coin's history with timestamped entries:

- **Event Types**:
  - "Acquired" (automatic on creation)
  - "Cleaned"
  - "Sent for grading" / "Received from grading"
  - "Displayed at show"
  - "Restored"
  - Custom entries

- **Add Entries** — Click the plus icon in Activity section
- **Delete Entries** — Remove entries you no longer need

## Market Value

When the coin is linked to a Numista type, the **Market value (Numista)** section shows Numista's community estimates for each issue and grade, in the collection currency, with your grade highlighted. It opens on the item's own issue when one was chosen (see [Banknote Details](#banknote-details)), else on the issue of its year.

## Sets

The **Sets** section lists the sets the coin belongs to. Add the coin to a set, take it out of one, or accept a suggested set.

## Wish List

- **Wishlist** — Coins you'd like to acquire stay out of the collection totals and stats
- **Mark as purchased** — Moves a wishlist coin into the collection, with the price, date and place of purchase

## History & Versioning

- **Last Updated** — Timestamp of last modification
- **Update Log** — View past changes (when available)

## Related Features

- See [AI Coin Analysis](ai-analysis.md) for vision-model analysis
- [Numista Catalog Lookup](numista-integration.md) for catalog references
- [Coin Sets](coin-sets.md) to group coins thematically

## API Endpoints

```
GET    /api/coins/:id               # Get coin details with all preloaded associations
PUT    /api/coins/:id               # Update coin fields
POST   /api/coins/:id/analyze       # Analyze coin images
DELETE /api/coins/:id/analyze       # Delete saved AI analysis
GET    /api/coins/:id/references    # List catalog references
POST   /api/coins/:id/references    # Add catalog reference
PUT    /api/coins/:id/references/:referenceId # Update reference
DELETE /api/coins/:id/references/:referenceId # Delete reference
GET    /api/coins/:id/journal       # List journal entries
POST   /api/coins/:id/journal       # Add activity journal entry
DELETE /api/coins/:id/journal/:entryId # Delete journal entry
```

## See Also

- [Collection Management](collection-management.md) — Browse and organize coins
- [Collection Statistics](statistics.md) — Analytics across your collection
- [PDF Export](pdf-export.md) — Export collection with full coin details
