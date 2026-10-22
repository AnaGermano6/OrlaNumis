# Ancient Coins — Web Frontend

Vue 3 + TypeScript + Pinia + Vite PWA for managing a personal ancient coin collection. Dark-themed, mobile-friendly single-page application that communicates with the Go API backend.

## Prerequisites

- **Node.js** `^20.19.0` or `>=22.12.0`

## Install & Run

```bash
npm install          # install dependencies
npm run dev          # start dev server (http://localhost:5173, proxies /api to :8080)
npm run build        # type-check + production build
npm run preview      # preview production build locally
```

## Testing

```bash
npm run test         # run tests with Vitest
npm run test:watch   # run tests in watch mode
npm run type-check   # vue-tsc type checking
```

## Key Dependencies

| Dependency | Purpose |
|---|---|
| `vue` 3.5+ | UI framework (Composition API, `<script setup>`) |
| `vue-router` 5 | Client-side routing |
| `pinia` 3 | State management |
| `axios` | HTTP client (JWT interceptor, 401 refresh queue) |
| `vite` 7 | Build tool and dev server |
| `vite-plugin-pwa` | Service worker and PWA manifest generation |
| `lucide-vue-next` | Icon library |
| `markdown-it` + `dompurify` | Render and sanitize markdown (AI chat) |
| `sortablejs` | Drag-and-drop reordering |
| `@imgly/background-removal` | Client-side image background removal |
| `vitest` | Unit testing framework |
| `vue-tsc` | TypeScript checking for `.vue` files |

## Directory Structure

```
src/
  pages/             # Route-level page components
  components/        # Reusable UI components
  stores/            # Pinia stores (auth, coins, settings, etc.)
  api/               # Axios client with JWT interceptor (client.ts)
  composables/       # Vue composables (shared reactive logic)
  utils/             # Pure utility functions
  types/             # TypeScript type definitions
  assets/styles/     # CSS variables and global styles
  router/            # Vue Router configuration
  __tests__/         # Vitest test files
```

## Routes

| Path | Page | Auth |
|---|---|---|
| `/login` | LoginPage | No |
| `/register` | RegisterPage | No |
| `/` | CollectionPage (coin gallery) | Yes |
| `/collection` | CollectionOverviewPage | Yes |
| `/banknotes` | BanknotesPage | Yes |
| `/banknotes/add` | AddBanknotePage | Yes |
| `/coin/:id` | CoinDetailPage | Yes |
| `/coin/:id/journal` | CoinDetailJournalPage | Yes |
| `/coin/:id/health` | CoinDetailHealthPage | Yes |
| `/coin/:id/notes` | CoinDetailNotesPage | Yes |
| `/coin/:id/actions` | CoinDetailActionsPage | Yes |
| `/coin/:id/analysis` | CoinDetailAnalysisPage | Yes |
| `/coin/:id/valuation` | CoinDetailValuationPage | Yes |
| `/add` | AddCoinPage | Yes |
| `/import/ucoin` | UcoinImportPage | Yes |
| `/quick-capture` | QuickCapturePage | Yes |
| `/quick-capture/drafts` | QuickCaptureDraftsPage | Yes |
| `/quick-capture/drafts/:id` | QuickCaptureDraftPage | Yes |
| `/lookup` | CoinLookupPage | Yes |
| `/deep-analysis/history` | DeepAnalysisHistoryPage | Yes |
| `/deep-analysis/:jobId?` | DeepAnalysisPage | Yes |
| `/edit/:id` | EditCoinPage | Yes |
| `/wishlist` | WishlistPage | Yes |
| `/wishlist/search-alerts` | WishlistAlertsPage | Yes |
| `/wishlist/availability-runs` | WishlistAvailabilityHistoryPage | Yes |
| `/wishlist/availability-runs/:id` | WishlistAvailabilityHistoryPage | Yes |
| `/stats` | StatsPage | Yes |
| `/stats/mint-map` | CollectionMapPage | Yes |
| `/stats/timeline` | TimelinePage | Yes |
| `/stats/time-machine` | StatsTimeMachinePage | Yes |
| `/stats/health` | StatsHealthPage | Yes |
| `/stats/distribution` | CollectionDistributionPage | Yes |
| `/mint-map` | Redirects to `/stats/mint-map` | — |
| `/timeline` | Redirects to `/stats/timeline` | — |
| `/settings` | SettingsPage | Yes |
| `/admin` | AdminPage | Yes + Admin |
| `/process-image` | Redirects to `/settings?tab=process` | — |
| `/sets` | SetsPage | Yes |
| `/sets/commemoratives` | CommemorativeCataloguePage | Yes |
| `/sets/:id` | SetDetailPage | Yes |
| `/sets/:id/info` | SetInsightsPage | Yes |
| any other path | Redirects to `/` | — |

A global navigation guard redirects unauthenticated users to `/login` and non-admin users away from `/admin`.

## PWA

The app uses `vite-plugin-pwa` with `registerType: 'prompt'`:

- **Manifest**: `standalone` display mode, dark theme (`#1a1a2e` / `#0f0f1a`)
- **Precaching**: All JS, CSS, HTML, images, and fonts via Workbox glob patterns
- **Runtime caching**: none. API reads and writes go straight to the network; `/uploads/*` is not runtime cached, and legacy `coin-images` caches are cleared on logout or user switch
- **Navigation fallback**: Denies `/api`, `/uploads`, and `/sw.js`
- **Update notification**: New service workers wait until the user accepts the in-app update banner; update checks run on registration, hourly, and when the app returns to the foreground

## Design System

All styling uses design tokens defined in `assets/styles/variables.css` and global utility classes from `main.css`. Key tokens include `--accent-gold`, `--bg-card`, `--border-subtle`, `--text-primary`, and `--radius-sm` through `--radius-full`. See [docs/design-system.md](../../docs/design-system.md) for the full token reference and component classes (`.chip`, `.btn`, `.badge`, etc.).
