# devnews

A terminal dashboard that aggregates engineering blog posts from top tech companies via RSS. Run `devnews` and browse the latest posts from Cloudflare, GitHub, Stripe, Netflix and more — without leaving your terminal.

```
┌─────────────────────────────────────────────────────────────┐
│  devnews — Engineering Blog Aggregator           Feb 23    │
├─────────────────────────────────────────────────────────────┤
│ [All] [Cloudflare] [GitHub] [Stripe] [Netflix] ...         │
├──────────────────────┬──────────────────────────────────────┤
│                      │                                      │
│ > Building DNS       │  Building DNS in Rust                │
│   Cloudflare · 2h    │                                      │
│                      │  We recently rewrote our DNS proxy   │
│   Scaling Idempot... │  in Rust for improved performance... │
│   Stripe · 5h        │                                      │
│                      │  Read more: https://blog.cloud...    │
│   JVM Tuning at...   │                                      │
│   Netflix · 1d       │                                      │
│                      │                                      │
├──────────────────────┴──────────────────────────────────────┤
│  12 articles · Cloudflare    ↑↓ navigate  o open  / search │
└─────────────────────────────────────────────────────────────┘
```

## Features

- **Morning briefing** — see what's new since your last visit: post count, most active sources, trending keywords
- **Reading streak** — tracks your daily usage streak in the status bar
- **AI summaries** — optional one-line summaries and topic tags via Claude or OpenAI
- **Two-pane layout** — article list + preview side by side
- **Source filtering** — toggle sources on/off with a tab bar
- **Search** — filter articles by title or description
- **SQLite cache** — instant startup after first fetch
- **Adaptive colors** — looks good in both dark and light terminals
- **Open in browser** — press `o` to read the full article
- **Zero config** — works out of the box, customizable via YAML

## Install

### Homebrew (macOS/Linux)

```bash
brew tap matheuskafuri/devnews
brew install devnews
```

### Go

Requires Go 1.21+:

```bash
go install github.com/matheuskafuri/devnews@latest
```

### Binary download

Download pre-built binaries for your platform from [GitHub Releases](https://github.com/matheuskafuri/devnews/releases). Available for Linux, macOS, and Windows on both amd64 and arm64.

```bash
# Example for macOS arm64
tar -xzf devnews_*_darwin_arm64.tar.gz
sudo mv devnews /usr/local/bin/
```

### Build from source

```bash
git clone https://github.com/matheuskafuri/devnews.git
cd devnews
make build
./devnews
```

## Usage

```bash
devnews                          # launch TUI
devnews --since 7d               # only show articles from last 7 days
devnews --since 24h              # only show articles from last 24 hours
devnews --refresh                # force refresh feeds before launching
devnews --config path/to/file    # use a custom config file
devnews stats                    # show cache size and article count
devnews prune                    # delete articles older than retention period
devnews prune --older-than 30d   # delete articles older than 30 days
devnews version                  # print version info
```

On first run, devnews fetches all configured feeds and caches them locally in SQLite. Subsequent launches load from cache instantly and only re-fetch when the refresh interval has elapsed (default: 1 hour).

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` or `↑` / `↓` | Move up/down in article list |
| `tab` | Switch focus between list and preview panes |
| `j` / `k` (in preview) | Scroll preview content |

### Actions

| Key | Action |
|-----|--------|
| `o` or `enter` | Open selected article in your default browser |
| `r` | Refresh all feeds |
| `/` | Enter search mode — filter by title or description |
| `f` | Enter filter mode — toggle sources on/off |

### Filter mode

| Key | Action |
|-----|--------|
| `←` / `→` or `h` / `l` | Move between sources |
| `space` or `enter` | Toggle selected source |
| `1`-`9` | Toggle source by number |
| `esc` or `f` | Exit filter mode |

### General

| Key | Action |
|-----|--------|
| `?` | Toggle help overlay |
| `esc` | Exit search/filter mode |
| `q` or `ctrl+c` | Quit |

## Configuration

On first run, a default config is written to `~/.config/devnews/config.yaml` (follows [XDG Base Directory](https://specifications.freedesktop.org/basedir-spec/latest/) on all platforms).

```yaml
refresh_interval: 1h
retention: 90d
sources:
  - name: Cloudflare
    type: rss
    url: https://blog.cloudflare.com/rss
    enabled: true
  - name: GitHub
    type: rss
    url: https://github.blog/engineering/feed/
    enabled: true
  - name: Stripe
    type: rss
    url: https://stripe.com/blog/feed.rss
    enabled: true
  - name: Netflix
    type: rss
    url: https://netflixtechblog.com/feed
    enabled: true
```

### Adding your own sources

Add any RSS or Atom feed to the `sources` list:

```yaml
sources:
  # ... existing sources ...
  - name: My Company
    type: rss
    url: https://engineering.mycompany.com/feed.xml
    enabled: true
```

### Disabling a source

Set `enabled: false` to hide a source without removing it:

```yaml
  - name: Stripe
    type: rss
    url: https://stripe.com/blog/feed.rss
    enabled: false
```

### Changing refresh interval

```yaml
refresh_interval: 30m   # fetch new articles every 30 minutes
```

### Changing retention period

Articles older than the retention period are automatically deleted after each feed refresh. Default is 90 days.

```yaml
retention: 30d   # keep only the last 30 days of articles
```

### AI summaries (optional)

Add an `ai` block to your config to enable one-line article summaries, topic tags, and a TL;DR briefing line. This is fully optional — devnews works great without it.

```yaml
ai:
  provider: claude          # claude | openai
  api_key: sk-ant-...       # or set DEVNEWS_AI_KEY env var
  model: claude-haiku-4-5-20251001  # optional, defaults to a fast model
```

You can also set the API key via environment variable instead of the config file:

```bash
export DEVNEWS_AI_KEY=sk-ant-...
```

When enabled:
- **Article summaries** — a one-line summary appears in the preview pane (generated on selection, cached in SQLite)
- **Topic tags** — up to 3 tags per article shown in the list and preview
- **TL;DR briefing** — AI-generated "why it matters" summaries on briefing cards and detected themes on the opening screen

## Default sources

| Source | URL |
|--------|-----|
| Cloudflare | https://blog.cloudflare.com/rss |
| GitHub | https://github.blog/engineering/feed/ |
| Stripe | https://stripe.com/blog/feed.rss |
| Netflix | https://netflixtechblog.com/feed |
| Meta | https://engineering.fb.com/feed/ |
| Spotify | https://engineering.atspotify.com/feed/ |
| Slack | https://slack.engineering/feed/ |
| Vercel | https://vercel.com/atom |
| Figma | https://www.figma.com/blog/feed/atom.xml |
| Lyft | https://eng.lyft.com/feed |

## Storage

devnews caches articles in a local SQLite database at `~/.cache/devnews/devnews.db` (XDG-compliant).

**Auto-pruning**: after each feed refresh, articles older than the `retention` period (default: 90 days) are automatically deleted and the database is vacuumed to reclaim disk space.

**Manual management**:

```bash
# Check how many articles are cached and how much space they use
devnews stats

# Delete articles older than the configured retention period
devnews prune

# Delete articles older than a specific duration
devnews prune --older-than 14d
```

With 8 default sources, the database typically stays under 200 KB.

## How it works

1. **Fetch** — devnews concurrently fetches RSS/Atom feeds from all enabled sources
2. **Cache** — articles are stored in a local SQLite database (see [Storage](#storage))
3. **Prune** — old articles are automatically deleted after each refresh based on the retention period
4. **Display** — a bubbletea TUI renders a two-pane interface with list + preview
5. **Refresh** — feeds are re-fetched when the configured interval has elapsed, or on demand with `r` or `--refresh`

No CGo required — the SQLite driver is pure Go (`modernc.org/sqlite`), so the binary is fully self-contained and works on any platform without external dependencies.

## Morning briefing

On launch, devnews shows a briefing header with:

- A time-of-day greeting
- How many new posts arrived since your last visit
- The most active sources
- Trending keywords (via TF-IDF) or a TL;DR line (if AI is configured)

Press any key to dismiss and enter the normal two-pane view. The briefing only appears when there are new articles since your last session.

## Reading streak

devnews tracks a daily reading streak. Open it every day and watch your streak grow in the status bar. Skip a day and it resets to 1. No gamification beyond the counter — just a quiet number.

## Development

```bash
# Build
make build

# Run
make run

# Test
make test

# Lint (requires golangci-lint)
make lint

# Clean
make clean
```

## License

[MIT](LICENSE)
