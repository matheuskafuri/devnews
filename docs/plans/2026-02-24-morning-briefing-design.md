# Morning Briefing — Design Doc

**Date:** 2026-02-24
**Status:** Approved

## Summary

Transform devnews from a passive RSS reader into a terminal-native engineering intelligence briefing. Three features, layered by complexity, all opt-in or zero-config.

## Feature 1: Morning Briefing Header

A dismissible header shown on launch that tells the user what's new.

### Components

- **Dynamic greeting**: "Good morning" / "Good afternoon" / "Good evening" based on local time
- **New post count**: Articles fetched since `last_opened` timestamp (stored in SQLite `meta` table)
- **Most active sources**: Top 2-3 sources by article count since last visit, shown as "Cloudflare (5), GitHub (3)"
- **Trending keywords**: Top 3 keywords extracted via local TF-IDF heuristic across new article titles — no LLM needed
- **Dismissible**: Press any key to collapse the header and enter the normal two-pane view

### Layout

```
┌─────────────────────────────────────────────────────────────┐
│  Good morning, engineer                          Feb 24    │
│  14 new posts since yesterday                              │
│  Most active: Cloudflare (5), GitHub (3)                   │
│  Trending: rust, dns, scaling                              │
│                                              [any key] ▸   │
├──────────────────────────────────────────────────────────────┤
│ ... normal TUI below ...                                    │
```

### Data

- `meta` table: `last_opened` key stores ISO timestamp, updated on each launch
- TF-IDF: score = (term frequency in new titles) / (number of articles containing term across all cached articles). Filter out stop words and terms < 4 chars.

## Feature 2: Reading Streaks

A quiet streak counter in the status bar. No gamification beyond the number.

### Logic

- Track `streak_days` (int) and `last_active_date` (YYYY-MM-DD) in `meta` table
- On launch:
  - If `last_active_date` == today: no-op
  - If `last_active_date` == yesterday: increment `streak_days`
  - Otherwise: reset `streak_days` to 1
- Update `last_active_date` to today

### Display

```
│  12 articles · All sources · streak: 7d    ↑↓ o / f ? q   │
```

### Edge cases

- First launch: streak = 1
- Skip a day: reset to 1 (not 0 — you showed up today)
- Multiple launches same day: no double counting

## Feature 3: Optional LLM Summaries

Opt-in intelligence features unlocked by adding an `ai:` block to config. Without it, devnews works exactly as before.

### Config

```yaml
ai:
  provider: claude          # claude | openai
  api_key: sk-ant-...       # or set DEVNEWS_AI_KEY env var
  model: claude-haiku-4-5-20251001  # optional, sensible defaults per provider
```

### Sub-features

**A) One-line article summaries**
- Generated lazily when an article is selected in the preview pane
- Displayed with `░` prefix above the description
- Cached in a `summary` column on the `articles` table (never re-generated)
- Prompt: "Summarize this engineering blog post in one sentence (max 120 chars): {title} — {description}"

**B) Topic tags**
- Up to 3 tags per article (e.g., `infrastructure`, `rust`, `performance`)
- Generated alongside the summary in the same LLM call
- Cached in a `tags` column (comma-separated)
- Displayed as dimmed labels next to the source name in the article list

**C) TL;DR briefing**
- Replaces the "Trending" line in the morning briefing header
- One sentence summarizing the day's new articles
- Generated once per session on launch (only if there are new articles)
- Prompt: "In one sentence, summarize the themes across these {n} engineering blog posts: {titles}"

### Provider implementation

- Simple HTTP client — no SDK dependencies
- Interface: `type Summarizer interface { Summarize(ctx, title, description string) (summary string, tags []string, err error) }`
- Two implementations: `claudeProvider` and `openaiProvider`
- Errors are non-fatal: if LLM fails, article displays normally without summary

## Codebase Changes

| Change | File |
|---|---|
| New | `internal/briefing/briefing.go` — header generation, TF-IDF, greeting |
| New | `internal/ai/ai.go` — Summarizer interface, Claude + OpenAI providers |
| Modify | `internal/cache/cache.go` — `last_opened`, streak meta keys, `summary`/`tags` columns |
| Modify | `internal/tui/app.go` — briefing header on launch, LLM summaries in preview |
| Modify | `internal/tui/statusbar.go` — streak counter |
| Modify | `internal/config/config.go` — optional `ai:` config block |
| Remove | `internal/summary/summary.go` — replaced by `internal/ai/` |

## Principles

- Zero-config by default: briefing + streaks work out of the box
- LLM features are fully opt-in: no API key = no AI, no errors
- No new dependencies beyond net/http for LLM API calls
- Same keybindings, same two-pane layout, same speed
- Non-fatal AI errors: if the LLM call fails, everything still works
