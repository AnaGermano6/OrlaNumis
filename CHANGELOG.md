# Changelog

Each version's section is its release notes: the release workflow publishes
it on GitHub, and **Admin → Updates** shows it to the apps that can update.
Versions follow [Semantic Versioning](https://semver.org).

## [1.0.0] - 2026-10-23

The first public release of Orla Numis, a self-hosted catalogue for a
personal collection of coins and banknotes, based on
[Aurearia](https://github.com/briandenicola/Aurearia).

### Collection
- Coin and banknote galleries with search and filters (face value, country
  with flags, period, coin type, presentation, rarity, set), viewed by
  country, period, monetary system, year, rarity or presentation.
- Coin and banknote records: numismatic details, photos, catalogue
  references (KM, Y, Pick, Numista), the period, worked out by the app,
  value history, previous and next. Banknotes also have their serial
  number, signatures and variant.
- Official photos from the ECB, INCM, Monnaie de Paris and Münze
  Österreich.
- Sets: the commemorative catalogue, challenges (each euro country, joint
  issues, a year's series, Portuguese escudo and réis types), filters, free
  sets and goals, with completion by country.
- Wish list, duplicates to trade, and lists to share as PDF or CSV; export
  to Excel or CSV; collection health that suits modern coins and
  banknotes.

### Numista
- Catalogue lookups and coins added from the catalogue; records completed
  from the cached types; values updated in daily batches within the monthly
  quota, with the requests counted by kind.

### AI (optional)
- Photo analysis, and a chat agent that searches European dealers of
  modern coins and banknotes. Anthropic Claude or a self-hosted Ollama.

### Running it
- One Docker image with the app and its AI agent, for amd64 and arm64:
  `anagermano/orlanumis`. Data in three folders; PUID and PGID for NAS
  systems; secrets made on the first start.
- Admin → Updates tells when a new version is out, with its notes, makes a
  backup and updates through Watchtower when one is set up.
- A full backup of the database and the photos every day, with import and
  restore.
- An Unraid template.
