# Changelog

Each version's section is its release notes: the release workflow publishes
it on GitHub, and **Admin → Updates** shows it to the apps that can update.
Versions follow [Semantic Versioning](https://semver.org).

## [1.0.0] - Unreleased

The first public release of Orla Numis, a self-hosted catalogue for a
personal collection of coins and banknotes, based on
[Aurearia](https://github.com/briandenicola/Aurearia).

### Collection
- Coin and banknote galleries with search and filters (face value, country
  with flags, period, coin type, presentation, rarity, set), viewed by
  country, period, monetary system, year, rarity or presentation.
- Coin and banknote records: numismatic details, photos, catalog references
  (KM, Y, Pick, Numista), periods worked out from Numista and uCoin, value
  history, previous and next.
- Sets: the commemorative catalogue, challenges (each euro country, joint
  issues, a year's series, Portuguese escudo and réis types), filters, free
  sets and goals, with completion by country.
- Wish list, duplicates to trade, and lists to share as PDF or CSV; export
  to Excel or CSV.

### Numista
- Catalogue lookups and coins added from the catalogue; records completed
  from the cached types; values updated in daily batches within the monthly
  quota, with the requests counted by kind.

### Running it
- One Docker image with the app and its AI agent, for amd64 and arm64:
  `anagermano/orlanumis`. Data in three folders; PUID and PGID for NAS
  systems; secrets made on the first start.
- Admin → Updates tells when a new version is out, with its notes, makes a
  backup and updates through Watchtower when one is set up.
- An Unraid template.
