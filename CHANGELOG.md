# Changelog

Each version's section is its release notes: the release workflow publishes
it on GitHub, and **Admin → Updates** shows it to the apps that can update.
Versions follow [Semantic Versioning](https://semver.org).

## [1.1.1] - 2026-11-02

### Changed
- The banknote gallery on a phone is laid out as the coin gallery: only the
  search stays at the top, with the filters button (showing how many filters
  are on) and a button to add a banknote; the order, grouping, country, kind
  of banknote, grid or list, selection and export open in a sheet; two
  banknotes a row, in compact cards.

## [1.1.0] - 2026-11-01

### Changed
- On a phone the app has one layout, the same in the browser and installed,
  chosen by the size of the screen (also when the phone is turned sideways).
- The gallery on a phone: only the search stays at the top; the filters, the
  order and the view open in a sheet from a button that shows how many
  filters are on; two coins a row, in compact cards; one choice of view
  (grid, list or swipe).
- After signing in, the app opens the collection overview.
- Every dialog rises from the bottom on a phone, at most 90% of the screen
  high and scrolling on its own, and is centred on a wider screen. The page
  behind stays still, and Escape or a tap outside closes it.
- The search for a coin's real photos offers only photos of a coin of its
  country, of its face value or year, the same face value first. Wikimedia
  Commons is searched within the categories of that country's coins first.
  Photos where a coin only shows the size of something else, and unrelated
  ones, are left out; when Commons has none of the coin, none is offered.

### Fixed
- Panels that went past the bottom of the screen and could not be reached,
  such as the gallery's filters in the installed app.
- The agent's button showing over dialogs; it hides while one is open.
- Pages wider than the screen on a phone (the gallery, adding a coin).
- iOS zooming into a field when it is tapped: fields are 16px on a phone.
- Titles cut short beside their buttons (the wishlist, a coin's page).
- Two prompts at once on the first start: the install prompt waits for the
  getting-started one.

## [1.0.3] - 2026-11-01

### Fixed
- The search for a coin's real photos works for coins from before 1999: it
  refused them, and said photos could not be searched right now.

### Changed
- Real photos are found on Wikimedia Commons more often: it is searched in
  English, as Commons names coin photos ("200 lire San Marino 1995 coin"),
  for images only, with the coin's year and then without it, before the
  words typed. A text typed instead of the coin's name is searched as it is.
- Numista does not let the app save its photos, so a Numista photo now opens
  to be saved by hand, with a note saying so, instead of failing on saving.

## [1.0.2] - 2026-10-25

### Fixed
- An idle app uses a fraction of the processor it did, which matters on a
  small server:
  - the background queues (Coin Copilot, AI jobs, the set builder, search
    alerts) no longer query the database up to ten times a second; new work
    still starts at once, and missed work is looked for every five seconds;
  - the container's health check no longer starts Python every 30 seconds.
- Without AI set up, `AI_AGENT=off` also saves the AI agent's memory and
  processor time; the Unraid template says so.

## [1.0.1] - 2026-10-25

### Fixed
- The container starts when a setting is left empty, as Unraid passes it:
  an empty setting gets the image's default back. An empty trusted proxies
  setting stopped the start.
- The log no longer shows a "record not found" error, with its SQL, for
  each row a new install looks up before creating it.

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
