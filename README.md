# crypto

<p><a href="https://go.dev" target="_blank"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="go version" /></a>&nbsp;<a href="https://github.com/mrcnserkan/crypto/blob/master/LICENSE.md" target="_blank"><img src="https://img.shields.io/badge/license-MIT-red?style=for-the-badge&logo=none" alt="license" /></a>&nbsp;<a href="https://github.com/mrcnserkan/crypto/releases/tag/v2.0.0" target="_blank"><img src="https://img.shields.io/badge/version-v2.0.0-blue?style=for-the-badge&logo=none" alt="version" /></a></p>

<p>A powerful and user-friendly CLI tool for real-time cryptocurrency tracking, portfolio management, and market analysis.</p>

## Features

- Real-time cryptocurrency price tracking
- Terminal trading charts with line and candlestick modes (LTTB downsampling)
- OHLC period stats (Open, High, Low, Close, change %)
- Custom date range filtering for charts (UTC)
- Portfolio management with weighted-average P&L
- Price alerts with foreground watch and background daemon
- Watchlist for tracking favorite coins
- Portfolio export (CSV / JSON)
- Multi-currency support (USD, EUR, TRY, etc.)
- Config file support (`~/.crypto/config.json`)
- Shell completion (bash, zsh, fish)
- API resilience: retry on 429/5xx, rate limiting, optional API key

## Installation

### Prerequisites

- [Go 1.21+](https://golang.org/dl/) installed on your system
- Internet connection for real-time data fetching

### Quick Install

```bash
go install github.com/mrcnserkan/crypto@v2.0.0
```

### Build from Source

```bash
git clone https://github.com/mrcnserkan/crypto.git
cd crypto
go build
```

## Configuration

Settings are stored in `~/.crypto/config.json`. Priority: **CLI flag > config file > default**.

```json
{
  "currency": "usd",
  "chart_width": 80,
  "chart_height": 20,
  "alert_check_interval_minutes": 5,
  "no_color": false
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `COINGECKO_API_KEY` | CoinGecko Demo/Pro API key |
| `COINGECKO_API_KEY_HEADER` | Header name: `x-cg-demo-api-key` (default) or `x-cg-pro-api-key` |

## Usage

### Basic Commands

```bash
crypto -h
crypto                          # Top coins (page 1)
crypto --page 2
crypto --per-page 20
crypto --currency eur
crypto --no-color               # Disable colors globally
```

### Shell Completion

```bash
# bash
source <(crypto completion bash)

# zsh
source <(crypto completion zsh)

# fish
crypto completion fish | source
```

### Interactive Charts

```bash
crypto bitcoin --graph
crypto bitcoin --graph --candles
crypto bitcoin --graph --interval 30d
crypto bitcoin --graph --from 2026-06-01 --to 2026-06-30
crypto bitcoin --graph --width 100 --height 24
```

Available intervals: `1d`, `7d`, `14d`, `30d`, `90d`, `180d`, `1y`, `max`

Date filters are parsed in UTC. Chart output includes aligned Y/X axes and period OHLC summary.

### Portfolio Management

```bash
crypto portfolio add bitcoin 0.5 50000 buy
crypto portfolio add bitcoin 0.1 55000 sell
crypto portfolio list
crypto portfolio history
crypto portfolio export --format csv --output portfolio.csv
crypto portfolio export --format json
crypto portfolio remove bitcoin
crypto portfolio clear
```

Portfolio list includes **Avg Cost**, **P&L**, and **P&L %** columns using weighted-average cost basis.

> P&L uses weighted average and is for personal tracking only — not for tax or accounting.

Use the global `--currency` flag for valuation (not a portfolio-specific flag):

```bash
crypto portfolio list --currency eur
```

### Price Alerts

**Breaking change in v2.0:** Alerts are no longer checked automatically when you run other commands. You must start monitoring explicitly.

```bash
crypto alert add bitcoin 50000 above
crypto alert add bitcoin 45000 below
crypto alert list
crypto alert remove bitcoin
crypto alert remove bitcoin 50000 above   # Remove specific alert

# Foreground (blocks terminal, Ctrl+C to stop)
crypto alert watch

# Background daemon (macOS/Linux)
crypto alert start
crypto alert status
crypto alert stop
```

Daemon logs: `~/.crypto/alert.log` · PID file: `~/.crypto/alert.pid`

### Watchlist

```bash
crypto watchlist add bitcoin
crypto watchlist add ethereum
crypto watchlist list
crypto watchlist remove bitcoin
```

### Data Storage

All local data lives in `~/.crypto/`:

| File | Purpose |
|------|---------|
| `portfolio.json` | Holdings and transactions |
| `alerts.json` | Active price alerts |
| `watchlist.json` | Saved coin IDs |
| `config.json` | User preferences |
| `alert.pid` | Background daemon PID |

## Breaking Changes (v1.x → v2.0)

1. **Alerts** no longer auto-start — use `crypto alert watch` or `crypto alert start`
2. **`--currency`** is a single global flag — portfolio subcommand no longer defines its own
3. Portfolio list columns expanded (Avg Cost, P&L, P&L %)

## API Rate Limits

This tool uses CoinGecko's public API with built-in rate limiting (~10 req/min on free tier). Set `COINGECKO_API_KEY` for higher limits.

## Contributing

Contributions are welcome! Feel free to submit issues and pull requests.

## License

This project is licensed under the MIT License — see [LICENSE.md](LICENSE.md) for details.

## Acknowledgments

- Data provided by [CoinGecko](https://www.coingecko.com/)
- Built with [Go](https://golang.org/)
