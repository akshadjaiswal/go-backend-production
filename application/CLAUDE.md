# CLAUDE.md — Go Backend Production Documentation App

This file gives Claude the context needed to work effectively on this Next.js documentation app.

## What this app is

A statically generated documentation site built with Next.js 16. It reads Go backend learning content from the parent repository and presents them as a navigable, searchable documentation site. There are 10 stages total across 3 groups.

The app lives at: `go-backend-production/application/`
The parent repo root (content source) is: `go-backend-production/` (one level up via `process.cwd() + '/...'`)

## Content source — read this carefully

Content is NOT stored inside `application/`. It is read at build time from the parent repository:

| Stage dirs | Source | How parsed |
|-----------|--------|------------|
| stage-01-basics through stage-10-deployment | `../stage-NN-topic/README.md` | Directory scan — each directory is one stage page |

The entire parsing logic lives in `lib/stages.ts`. Do NOT move content files into `application/`.

## Directory naming convention

Stage directories follow a consistent pattern — no fallback needed:

```
stage-01-basics/
stage-02-routing/
stage-03-middleware/
stage-04-database/
stage-05-auth/
stage-06-validation/
stage-07-config/
stage-08-logging/
stage-09-testing/
stage-10-deployment/
```

Single regex: `/^stage-(\d+)-(.+)$/`

Slug format: strip the `stage-` prefix → `01-basics`, `02-routing`, ..., `10-deployment`

## Architecture

```
application/
├── app/
│   ├── layout.tsx              # Root layout with fonts, theme script, sidebar
│   ├── page.tsx                # Home page with 3 group sections + stage cards
│   ├── globals.css             # Tailwind base + Shiki styles + print CSS
│   ├── error.tsx               # Error boundary
│   ├── robots.ts               # robots.txt
│   ├── sitemap.ts              # Dynamic sitemap from getAllStages()
│   ├── opengraph-image.tsx     # Site-level OG image
│   └── stages/[slug]/
│       ├── page.tsx            # Stage page with TOC + nav + reading progress
│       ├── loading.tsx         # Skeleton loading state
│       ├── not-found.tsx       # 404 for bad slugs
│       └── opengraph-image.tsx # Per-stage OG image (1200×630, Go blue)
├── components/
│   ├── sidebar.tsx             # RSC wrapper — calls getGroups()
│   ├── sidebar-client.tsx      # Client: collapsible nav, mobile overlay, completion, progress
│   ├── header.tsx              # Sticky header with search, shortcuts, theme, GitHub link
│   ├── markdown-renderer.tsx   # Async RSC: Shiki dual-theme + callouts + HeadingAnchor
│   ├── table-of-contents.tsx   # Client: sticky TOC with scroll % — aware of docs-pane scroll root
│   ├── stage-nav.tsx           # Prev/Next stage links with read time
│   ├── stage-header.tsx        # RSC: stage title + metadata + bookmark/complete/print buttons
│   ├── split-layout.tsx        # Client: docs+code side-by-side on lg+, stacked on mobile
│   ├── code-pane.tsx           # Async RSC: highlights all stage source files with shiki
│   ├── code-pane-client.tsx    # Client: VS Code-style UI — tabs, file tree, code area, status bar
│   ├── code-explorer.tsx       # Async RSC: simpler code viewer (used standalone if needed)
│   ├── code-explorer-client.tsx # Client: directory tabs + file list + code viewer
│   ├── copy-button.tsx         # Copy code button for code blocks
│   ├── reading-progress.tsx    # Blue progress bar at top of page
│   ├── search-modal.tsx        # Fuse.js search modal
│   ├── search-trigger.tsx      # Search button + "/" key listener
│   ├── bookmark-button.tsx     # Bookmark toggle for stages
│   ├── continue-reading.tsx    # Sidebar continue reading / bookmark link
│   ├── theme-provider.tsx      # Dark mode context (gbp_theme localStorage key)
│   ├── theme-toggle.tsx        # Sun/Moon icon button
│   ├── heading-anchor.tsx      # Copy-link-to-section on headings
│   ├── complete-button.tsx     # Mark stage complete + group celebration toast
│   ├── stage-completion-badge.tsx  # Check icon on home page stage cards
│   ├── shortcuts-modal.tsx     # Keyboard shortcuts reference
│   ├── shortcuts-trigger.tsx   # "?" key listener + Keyboard icon button
│   ├── stage-shortcuts.tsx     # "b" key → bookmark (stage pages only)
│   ├── print-button.tsx        # window.print()
│   └── scroll-to-top.tsx       # Floating ArrowUp button
├── hooks/
│   └── use-bookmark.ts         # localStorage hooks with gbp_ prefix
├── lib/
│   ├── stages.ts               # ALL content parsing
│   ├── github.ts               # GitHub star count (cached 1h)
│   └── utils.ts                # cn() utility
├── types/
│   └── stage.ts                # Stage, StageMeta, Group, TocHeading interfaces
├── scripts/
│   ├── copy-stage-images.mjs   # Copies images from stage dirs to public/stage-images/
│   └── generate-search-index.mjs  # Writes public/search-index.json
└── public/
    ├── icon.svg
    ├── fonts/PlayfairDisplay.ttf   # Required for OG image generation — DO NOT DELETE
    ├── stage-images/               # Auto-generated — do not edit manually
    └── search-index.json           # Auto-generated — do not edit manually
```

## Key exported functions from `lib/stages.ts`

- `getAllStages(): StageMeta[]` — flat list of all 10 stages in order
- `getStageBySlug(slug: string): Stage | null` — full stage with content + headings + readTime
- `getGroups(): Group[]` — 3 groups with nested stage lists
- `dirNameToSlug(dirName: string): string` — strips `stage-` prefix
- `getSearchIndex()` — used by generate-search-index.mjs
- `getStageFiles(dirName: string): StageFileGroup[]` — reads all source files for a stage (`.go`, `.sql`, `.http`, `.env.example`, `.yml`, `Dockerfile`, `.dockerignore`); sorted by directory then by name with `main.go` pinned first

## Groups

| Group | Number | Stages | Description |
|-------|--------|--------|-------------|
| Foundations | 1 | 01-03 | HTTP, routing, middleware |
| Data & Auth | 2 | 04-06 | database, JWT auth, validation |
| Production | 3 | 07-10 | config, logging, testing, deployment |

## Design system

- **Framework**: Tailwind CSS 3 with custom config in `tailwind.config.js`
- **Accent color**: `#00ADD8` (Go's official blue)
- **Background**: `#FFFFFF`, **Foreground**: `#000000`
- **Fonts** (via `next/font/google`):
  - Heading: Playfair Display (`--font-heading`)
  - Body: Source Serif 4 (`--font-body`)
  - Mono: JetBrains Mono (`--font-mono`)
- **No border-radius, no box-shadow** — everything is sharp-cornered by design
- **Dark mode**: class-based with localStorage sync; blocking inline script prevents FOUC

## Stage page layout (split view)

On `lg+` (≥1024px) the stage page renders a **side-by-side split layout**:
- **Left pane** (`id="docs-pane"`): rendered README, TOC (xl+), stage nav — scrolls independently
- **Right pane**: VS Code-style code explorer (always dark) — sticky, fills viewport height
- **Drag handle**: 4px divider between panes, draggable to resize, ratio persisted in localStorage

On mobile (< lg): stacked — docs first, then code explorer below.

The `SplitLayout` client component owns the drag/resize logic. The `CodePane` RSC pre-highlights all source files with shiki at build time and passes them to `CodePaneClient` which renders the editor UI.

### CSS variable
```css
:root { --header-h: 53px; }  /* height of sticky header — used for calc(100vh - var(--header-h)) */
```

### Code pane always-dark
`.code-pane .shiki-light { display: none !important }` — in `globals.css`. The code pane is always dark regardless of theme. This is intentional (editor aesthetic).

### TOC scroll root
`TableOfContents` detects split mode by checking `document.getElementById('docs-pane')`. If found (clientWidth > 0), it uses the docs pane element as the `IntersectionObserver root` and scroll event target instead of `window`.

## localStorage keys

| Key | Purpose |
|-----|---------|
| `gbp_last_visited` | Last stage page visited |
| `gbp_bookmark` | Explicitly bookmarked stage |
| `gbp_completed` | JSON array of completed stage slugs |
| `gbp_theme` | `'light'` / `'dark'` / `'system'` |
| `gbp_split_ratio` | Docs/code pane width ratio (integer 25–75, default 55) |

## Data types

### `Stage` interface (`types/stage.ts`)
```ts
{
  slug: string
  dirName: string
  title: string
  number: string        // "01" through "10"
  group: 1 | 2 | 3
  groupLabel: string    // "Foundations" | "Data & Auth" | "Production"
  content: string
  headings: TocHeading[]
  readTime: number
}
```

`StageMeta = Omit<Stage, 'content' | 'headings'>`

## Build scripts

```bash
cd application
npm install
npm run dev       # Dev server at localhost:3000
npm run build     # Runs prebuild (images + search index) then Next.js build
npm start         # Serve production build
```

The `prebuild` script runs:
1. `node scripts/copy-stage-images.mjs` — copies images from stage dirs to public/stage-images/
2. `node scripts/generate-search-index.mjs` — writes public/search-index.json

## Data types

### `StageFile` + `StageFileGroup` (`types/stage.ts`)
```ts
interface StageFile {
  path: string      // relative: "handlers/auth.go"
  filename: string  // basename: "auth.go"
  dir: string       // "handlers" | "." for root-level files
  content: string   // raw text
  lang: string      // shiki language id: "go" | "sql" | "yaml" | "ini" | etc.
}

interface StageFileGroup {
  dir: string
  files: StageFile[]
}
```

### `HighlightedFile` (local to `code-pane.tsx` / `code-explorer.tsx`)
```ts
interface HighlightedFile {
  path: string; filename: string; dir: string; lang: string
  rawContent: string; lightHtml: string; darkHtml: string
  lineCount: number   // code-pane only — used for line number gutter
}
```

## Do not

- Do NOT move content into `application/` — source from parent repo at build time
- Do NOT add a database or CMS — all content is static
- Do NOT add rounded corners or box shadows — sharp corners by design
- Do NOT change the slug format — stable URLs (breaking them breaks bookmarks)
- Do NOT use `@shikijs/rehype` as a rehype plugin — use `codeToHtml()` directly
- Do NOT edit `public/search-index.json` or `public/stage-images/` manually — auto-generated
- Do NOT delete `public/fonts/PlayfairDisplay.ttf` — required for OG image generation
- Do NOT change `--header-h` in globals.css without also measuring the actual header height
- Do NOT expose real `.env` files in `getStageFiles()` — only `.env.example` and `.env.test.example`
- Do NOT use `@shikijs/rehype` as a rehype plugin — use `codeToHtml()` directly
