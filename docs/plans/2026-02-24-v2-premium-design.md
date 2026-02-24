# DevNews V2 — Premium Polished Edition

> DevNews — Your daily engineering intelligence briefing.

## Product Identity

DevNews V2 is a curated engineering briefing tool. Clean, confident, intentional. No gimmicks. The experience should feel like a private tool serious engineers use quietly — calm, in control, informed, efficient.

## Architecture

Two-mode system: briefing-first with browse escape hatch.

```
devnews                    # briefing mode (default)
devnews browse             # full article browser (current two-pane)
devnews --focus=infra      # briefing filtered to category
devnews --refresh          # force refresh before briefing
devnews stats              # cache stats (existing)
devnews prune              # manual prune (existing)
```

### Data Flow on Launch

1. Load articles from cache (auto-refresh if stale)
2. Score all articles via local signal engine
3. Select top 5 (or top 5 in focus category)
4. Classify each article by category (local keyword matching)
5. Generate "Why it matters" for uncached articles (AI, async)
6. Generate detected themes from top 5 (AI, async)
7. Render opening screen immediately (themes load in background)
8. User presses Enter — card-by-card navigation

## Signal Scoring Engine

**Package:** `internal/signal/`

Pure Go, no AI dependency. Four components, each normalized 0.0–1.0:

| Component | Weight | Logic |
|---|---|---|
| Recency | 0.30 | Exponential decay: 1.0 at publish, 0.5 at 24h, 0.1 at 72h |
| Source weight | 0.25 | Configurable per-source in config (default 0.5) |
| Depth | 0.25 | Description word count: <50w = 0.2, 50-150 = 0.6, 150+ = 1.0 |
| Keyword density | 0.20 | Match against curated high-signal engineering terms |

Final score: weighted sum, scaled to 0.0–10.0 for display.

Score breakdown accessible via `i` key on any card:

```
Signal Score Breakdown

Source weight:         0.95
Depth score:           0.82
Recency score:         0.90
Keyword density score: 0.78

Final: 8.7
```

Scores are computed at launch and stored in cache. Recomputed each session (recency changes).

## Category Classification

**Package:** `internal/classify/`

Fixed 7-category taxonomy with keyword-based matching. No AI.

| Category | Accent Color | Hex |
|---|---|---|
| AI/ML | Soft violet | #B39DDB |
| Infrastructure | Cool cyan | #80CBC4 |
| Databases | Subtle green | #A5D6A7 |
| Distributed Systems | Warm amber | #FFD54F |
| Security | Muted red | #EF9A9A |
| Developer Tools | Warm white | #D7CCC8 |
| Platform | Dim cyan | #80DEEA |

Classification logic:
1. Tokenize title + description (lowercase, strip punctuation)
2. Count keyword matches per category
3. Highest match count wins
4. Tie-break: first category in list order
5. Zero matches: defaults to Platform

Stored in cache as `category` column.

## Briefing UI

### Opening Screen

```
DevNews — Feb 24

Signal status: Fresh
Posts scanned: 38
Selected for briefing: 5

Detected themes:
• Reliability under scale
• AI inference performance
• JVM memory tuning

Press Enter to begin.
```

- "Signal status" — `Fresh` if refreshed this session, `Cached (2h ago)` if from cache
- "Posts scanned" — total article count in cache within retention window
- "Detected themes" — AI-generated (async), shows `Analyzing...` while loading, falls back to TF-IDF keywords if no AI
- Minimal whitespace-driven layout, no borders

### Card View

```

  1/5

  Stripe — Durable Idempotency in High-Volume Systems

  Category: Distributed Systems          Reading time: 4 min
  Signal: 8.7

  Why it matters:
  Introduces architectural patterns that reduce failure
  propagation across distributed payment workflows.

  ────────────────────────────────────────────────────────
  o open   n next   p previous   i info   a browse all   q quit

```

Card behavior:
- `n`/`p` (or `j`/`k`, arrow keys) — navigate between cards
- `o`/Enter — open in browser
- `i` — toggle signal score breakdown overlay
- `a` — exit to full browser mode
- `q` — quit
- No wrapping: `p` on card 1 does nothing, `n` on card 5 does nothing

### Focus Mode

```
devnews --focus=infra
```

Opening screen reflects the filter:

```
DevNews — Feb 24

Focus: Infrastructure

Posts scanned: 38
Infrastructure articles: 12
Selected for briefing: 5

Detected themes:
• Edge caching invalidation strategies
• Container runtime performance

Press Enter to begin.
```

Focus flag aliases:

| Flag | Category |
|---|---|
| infra | Infrastructure |
| distributed | Distributed Systems |
| ai | AI/ML |
| db | Databases |
| security | Security |
| tools | Developer Tools |
| platform | Platform |

Invalid flag shows clean error with valid options. Persistable in config as `focus: infra`.

## "Why It Matters" Layer

AI-generated, cached in SQLite. Only generated for the top 5 articles per briefing.

- Prompt style: measured, technical, precise. No hype.
- Bad: "This reduces blast radius!"
- Good: "Introduces architectural patterns that reduce failure propagation across distributed services."
- Cached in `why_it_matters` column — not regenerated on subsequent launches
- Falls back to first sentence of RSS description if no AI configured
- Shows `Preparing summary...` while generating

## Theme Detection

AI-generated from the top 5 article titles and descriptions.

- Returns 2-3 themes in analytical tone
- Cached per briefing session
- Falls back to TF-IDF keywords (existing implementation) if no AI
- Tone: "Resilience patterns in distributed systems" not "Big Tech Is Thinking About..."

## Reading Time

Estimated from description word count with multiplier (blog posts typically 3-5x their RSS summary). Displayed as whole minutes, minimum 1 min.

## Style System

Two style sets coexist:

### Briefing Styles (new)

| Role | Color | Hex |
|---|---|---|
| Title text | Bright white | #F0F0F0 |
| Body text | Light gray | #C0C0C0 |
| Metadata | Soft gray | #808080 |
| "Why it matters" | Warm gray | #A0A0A0 |
| Keybindings | Dim gray | #606060 |
| Separators | Subtle gray | #404040 |

- Accent colors only for category labels
- No bold except article titles
- No borders — whitespace and `─` separators only
- Generous padding

### Browse Styles (preserved)

Current styles unchanged:
- Bordered pane layout
- Tab-style filter bar
- Cyan accent (`#00E5FF` dark / `#0097A7` light)
- `░` prefix on AI summaries

## Config

```yaml
# ~/.config/devnews/config.yaml
brief_size: 5
refresh_interval: 12h
retention: 7d
focus: null
theme: default

sources:
  - name: Stripe
    url: https://stripe.com/blog/feed.rss
    type: rss
    weight: 0.9      # signal scoring weight (0.0-1.0, default 0.5)
  # ...

ai:
  provider: claude
  model: claude-haiku-4-5-20251001
  # api_key via DEVNEWS_AI_KEY env var
```

New fields: `brief_size`, `weight` on sources, `focus`. All backward-compatible.

## Cache Schema Changes

New columns on `articles` table:

```sql
signal_score   REAL DEFAULT 0,
category       TEXT DEFAULT '',
why_it_matters TEXT DEFAULT ''
```

## Package Changes

| Package | Change |
|---|---|
| `internal/signal/` | **New.** Signal scoring engine. |
| `internal/classify/` | **New.** Category classifier with keyword taxonomy. |
| `internal/tui/` | New briefing views (opening screen + cards). Browse mode preserved with current styles. |
| `internal/ai/` | New prompts: "Why it matters" generation, theme detection. |
| `internal/cache/` | New columns: signal_score, category, why_it_matters. |
| `internal/briefing/` | Reworked: orchestrates score → rank → classify → generate. |
| `cmd/` | `browse` subcommand, `--focus` flag, `brief_size` support. |

## Out of Scope

- No infinite scroll
- No social features (likes, comments, sharing)
- No plugins or experiments
- No new sources or feed types
- No onboarding flow
- No cluttered metrics beyond signal score
