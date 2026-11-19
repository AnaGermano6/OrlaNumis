# Changelog

Each version's section is its release notes: the release workflow publishes
it on GitHub, and **Admin → Updates** shows it to the apps that can update.
Versions follow [Semantic Versioning](https://semver.org).

## [1.5.0] - 2026-11-19

### Changed
- The app opens on Novos lançamentos: the installed app starts there, and so
  does the server's address alone; the gallery opens from the Coleção menu.
- The coin card: the photo in a light box with its grade in a circle, the
  name in capitals over "year · material · kind", and the quantity and value
  at the bottom; four coins in a row.
- The gallery's and the banknotes' search: the search, grid or list and the
  side in a card with the filters under them ("Mais filtros" for the rest),
  kept in sight while the coins scroll.
- The statistics: the condition of the pieces in each tab (the whole
  collection, the coins, the banknotes), "Classificar agora" opening the
  coins or the banknotes without a grade; the rulers by country, as
  monarchs or regimes; a line between the tabs and the rest.
- The wishlist: each coin side by side with its photo, its country and
  shop, its price (about its face value when none is saved), "Marcar como
  comprada" and a button to remove it; a filter by country and by year once
  it has more than six coins.
- The sets: each with a picture of its coins (three piled up for a set of
  different coins), a joint issue's year apart from its name, how many of
  its coins the collection has and a progress bar, two in a row, with a
  filter by status. By country: a card per country with its flag, sets,
  coins and progress, to search and sort, its sets listed when opened.
- A set's page: its short name with its year and kind, and one card with
  what it lists beside how far it is.
- The commemorative catalogue: the countries listed beside (to search; the
  euro's or the escudos), the country's progress, filters by year, value and
  status, and the coins in rows with a button to add them and a heart for
  the wishlist. The sets page no longer links to it (the menu does).
- Flags look the same everywhere, as in the coin gallery.

### Added
- Lista de Trocas, a page of its own under Coleção: the coins and banknotes
  held more than once, to search or narrow to a country, each with its
  total and "− 1 para troca +" to change how many there are; the list can
  be copied as text or shared as a file.
- Novos lançamentos lists the year's coins the ECB has not published yet:
  every euro country's €2 coins, and the collector coins of the countries
  whose mint the app does not read (San Marino's, Slovakia's, Germany's…),
  from Numista, without a picture. The commemorative catalogue lists the
  year's €2 coins too.

## [1.4.0] - 2026-11-17

### Changed
- On a wider screen the menu is in the top bar, each page with the pages
  under it in a list (Coleção, Identificar moeda, Agente, Estatísticas,
  Conjuntos); the account and the language are on the right. The wishlist
  is under Coleção; Estatísticas opens its own page first.
- The footer is gone: its credits are in Settings › Sobre, with the version.
- The collection's page is now Novos lançamentos: the €2 coins and the
  collector coins as tabs, with country, status, metal and edition filters
  beside them and an ⓘ that says where the list comes from.
- The statistics start with the collection at a glance, in tabs (Toda a
  coleção, Moedas, Notas): the pieces, the estimated value against the face
  value, the amount invested and the countries; the coins' tab adds the
  wishlist. The coins by Numista's kind of coin and the banknotes by
  circulation open their galleries filtered.
- The duplicates to trade are on the wishlist's page, for coins and for
  banknotes.
- Numbers always group their thousands: 3 874,51 € rather than 3874,51 €.
  A share that would round to 0% or 100% shows a decimal (0,3%).
- On the map, a country's coins scroll in a box as tall as the list of
  countries, under a header that stays.
- Saúde opens with a general diagnosis: the score as a gauge out of 100
  with its grade, each part as a bar (details, images, valuations, and AI
  analyses when an AI provider counts them), what each weighs, and a way to
  the coins that miss something of each part.

### Added
- Completion progress: how many of the €2 commemoratives and of the euro
  circulation coins (the eight of each year) the collection has, year by
  year, of every country or of one.
- Condition: every grade, best first, with its pieces and share, and those
  without a grade, which open the gallery to grade them.
- Completion by country, precious metals (the fine gold and silver, from
  the fineness Numista gives, plated coins left out) with the other
  materials, and the coins by decade.
- The gallery can show only the coins without a grade.
- Priority actions in Saúde: the coins to fix first, the most urgent
  first, each with its photo, what it misses most and the button that fixes
  it; of one country or of one thing missing, each choice with how many
  coins it would show. A quick review goes through them one coin at a time,
  with everything each one misses.

### Removed
- The time machine, the acquisition flow and the old summary cards of the
  statistics.

## [1.3.0] - 2026-11-17

### Changed
- The page of a coin (and of a banknote) reads as a museum label: its two
  sides on top, side by side and named, then its name and badges, then tabs
  so the page keeps one length: Detalhes, Inscrição, História, Referências
  de catálogo and Conjuntos.
- Detalhes opens first, with the material, weight, diameter and grade (a
  banknote's material, size and condition) in a row, and the other details
  under them. The tab chosen stays from coin to coin.
- História shows the coin's notes; Referências de catálogo has the Numista
  market value too.
- Weights and sizes are written in the language's numbers: "8,5 g",
  "25,75 mm".

### Added
- Conjuntos proposes up to three sets the coin could start, made in one
  click or opened when they exist: the catalogue album it belongs to (its
  country's €2 commemoratives or collector coins, its year's euro series,
  Portugal's escudos or réis), and sets that fill themselves with the
  collection's coins like it (its country's commemoratives, its year, its
  country). A banknote's are of banknotes.
- A set's filters can keep to coins or to banknotes.

### Fixed
- The suggested sets show only when there are some, without an empty
  "Ainda não há sugestões".

## [1.2.0] - 2026-11-16

### Added
- The collection overview shows the face value of the whole collection, of
  the coins and of the banknotes, and how much more than it they are worth
  ("Acima do facial"). It counts the face values in euros, and in escudos at
  their fixed rate of 200.482 escudos to the euro; the items in other
  currencies are counted apart ("179 sem valor facial"), and the percentage
  compares only the items that have both a face value and a value.
- New releases: filters by metal (gold, silver, base metal) and "Sem
  edições especiais", which hides the gold coins and those in proof or FDC
  finish. A design's gold, proof and FDC versions have cards of their own,
  so the filter keeps its others. The commemoratives catalogue has the same
  filter, for the designs struck only in gold, proof or FDC.
- The new releases' cards show the country's flag, and gold or silver beside
  the face value; the country filter shows the flags too.

### Changed
- "Novos lançamentos" replaces "Lançamentos", without the line under it.
- The €2 coins ("Moedas de 2 €") are the ECB's commemoratives of this year
  and the last, and the mints' €2 coins of a country and year the ECB does
  not list yet (it has not published its 2026 list); these were among the
  collector coins.
- France's and Austria's new coins say the month they go on sale, their
  edition ("Prata Proof", "Prata BNC") and their mintage, as Portugal's do.
  France's are read from each coin's page in the background: they fill in
  over a few minutes after an update.

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
