# TUI UX Overhaul — Design

## Overview

A comprehensive UX upgrade across five areas: list item redesign with at-a-glance intelligence, flexible pane layouts, rich preview pane, adaptive themes, and smooth navigation. The goal is to take devnews from functional to premium.

## Phase 1: List Item Redesign + At-a-Glance Intelligence

Each list item becomes a structured, scannable row:

```
  ● Cloudflare rewrote their DNS proxy in Rust     AI  2h
    Infrastructure
```

- **Read/unread indicator**: `●` (unread, bright) vs `○` (read, dim entire row)
- **Title**: full brightness if unread, dimmed if read
- **Category badge**: colored text using category color palette, shown below title
- **AI marker**: small `AI` tag (accent color) if `FullSummary` exists
- **Relative time**: color-coded — under 6h bright, under 24h normal, older dim
- **Selected item**: accent-colored `▸` prefix replaces the dot, full row highlight

Requires tracking read state per article (new `read` column in SQLite, set when user opens in browser or views in preview-only mode).

## Phase 2: Flexible Pane Layout

Three layout modes toggled with `v`:

| Mode | Layout | Use |
|------|--------|-----|
| **Split** (default) | 35% list / 65% preview | Browse + read |
| **List-only** | 100% list | Scan/triage |
| **Preview-only** | 100% preview | Deep read |

- `v` cycles: split → list-only → preview-only → split
- In list-only: `enter`/`o` opens browser, `v` jumps to preview-only
- In preview-only: `j`/`k` navigates articles, `v` returns to split
- Status bar shows layout: `◧ split`, `▯ list`, `▮ preview`

## Phase 3: Rich Preview Pane

Clear visual sections with separators:

- Category shown below source line in its category color
- AI Summary section with labeled header, visually separated by rules
- Loading state: animated progress indicator `░░░▒▒▒▓▓▓ Generating summary...`
- Preview-local hints at bottom: `S summarize  o open  tab focus  v layout`
- Scroll indicators: `▼ more` at bottom, `▲` at top when scrolled

## Phase 4: Adaptive Themes

4 built-in themes:

| Theme | Accent | Tone |
|-------|--------|------|
| **Neon** (default) | `#00E5FF` cyan | Dark |
| **Dracula** | `#BD93F9` purple | Dark |
| **Nord** | `#88C0D0` frost blue | Dark |
| **Solarized Light** | `#268BD2` blue | Light |

- `Theme` struct holding all color slots (accent, text, muted, dim, subtle, surface, body, category colors)
- Config: `theme: "neon"` in config.yaml
- `T` opens theme picker overlay (4 options, instant preview)
- Replaces `lipgloss.AdaptiveColor` with theme-driven colors

## Phase 5: Smooth Navigation

- Header gains layout indicator and breadcrumb trail: `Home › Browse`, `Home › Briefing › Card 3/5`
- Universal `esc`: always goes back one level
- `h` always means home (already mostly works)
- Cursor position preserved when switching modes
- Consistent mode transitions
