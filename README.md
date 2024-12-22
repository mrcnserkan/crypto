# crypto

<p><a href="https://go.dev" target="_blank"><img src="https://img.shields.io/badge/Go-1.17+-00ADD8?style=for-the-badge&logo=go" alt="go version" /></a>&nbsp;<a href="https://github.com/mrcnserkan/crypto/blob/master/LICENSE.md" target="_blank"><img src="https://img.shields.io/badge/license-MIT-red?style=for-the-badge&logo=none" alt="license" /></a>&nbsp;<a href="https://github.com/mrcnserkan/crypto/releases/tag/v1.2.2" target="_blank"><img src="https://img.shields.io/badge/version-v1.2.2-blue?style=for-the-badge&logo=none" alt="version" /></a></p>

<p>A powerful and user-friendly CLI tool for real-time cryptocurrency tracking, portfolio management, and market analysis.</p>

## Features

- 🚀 Real-time cryptocurrency price tracking
- 📊 Interactive line and candlestick charts
- 💼 Portfolio management with transaction history
- 🔔 Customizable price alerts
- 🌐 Multi-currency support (USD, EUR, TRY, etc.)
- 🔍 Advanced coin search functionality
- 📈 Detailed market statistics and trends

## Installation

### Prerequisites

- [Go 1.17+](https://golang.org/dl/) installed on your system
- Internet connection for real-time data fetching

### Quick Install

```bash
go install github.com/mrcnserkan/crypto@latest
```

### Build from Source

```bash
git clone https://github.com/mrcnserkan/crypto.git
cd crypto
go build
```

## Usage

### Basic Commands

```bash
# Display help information
crypto -h

# List top cryptocurrencies (default page: 1, per-page: 10)
crypto

# Navigate through pages
crypto --page 2

# Adjust results per page (max: 250)
crypto --per-page 20

# Change display currency (default: USD)
crypto --currency eur  # Supports USD, EUR, TRY, GBP, etc.
```

Example output:
```
🏆 Top Cryptocurrencies by Market Cap

  #  |       COIN        |       |  PRICE  |  24H  |   7D   | MARKET CAP |   ATH     
-----+-------------------+-------+---------+-------+--------+------------+-----------
  1  | Bitcoin           | BTC   | $95.49K | -1.4% | -7.4%  | $1.89T     | $108.14K
-----+-------------------+-------+---------+-------+--------+------------+-----------
  2  | Ethereum          | ETH   | $3.28K  | -1.6% | -15.8% | $396.08B   | $4.88K
-----+-------------------+-------+---------+-------+--------+------------+-----------
  3  | Tether            | USDT  | $1.00   | -0.2% | -0.2%  | $139.74B   | $1.32
-----+-------------------+-------+---------+-------+--------+------------+-----------
  4  | XRP               | XRP   | $2.22   | -0.0% | -8.1%  | $127.19B   | $3.40
-----+-------------------+-------+---------+-------+--------+------------+-----------
  5  | BNB               | BNB   | $651.05 | -0.5% | -9.4%  | $95.22B    | $788.84
-----+-------------------+-------+---------+-------+--------+------------+-----------
  6  | Solana            | SOL   | $179.47 | -0.2% | -18.8% | $86.15B    | $263.21
-----+-------------------+-------+---------+-------+--------+------------+-----------
  7  | Dogecoin          | DOGE  | $0.31   | -1.2% | -22.1% | $46.31B    | $0.73
-----+-------------------+-------+---------+-------+--------+------------+-----------
  8  | USDC              | USDC  | $1.00   | -0.2% | -0.1%  | $42.93B    | $1.17
-----+-------------------+-------+---------+-------+--------+------------+-----------
  9  | Lido Staked Ether | STETH | $3.28K  | -1.4% | -15.8% | $31.86B    | $4.83K
-----+-------------------+-------+---------+-------+--------+------------+-----------
  10 | Cardano           | ADA   | $0.89   | -0.7% | -18.7% | $31.81B    | $3.09
-----+-------------------+-------+---------+-------+--------+------------+-----------
```

### Detailed Coin Information

```bash
# View detailed information for a specific coin
crypto bitcoin

# Search for coins by name or symbol
crypto --search "solana"
```

Example output:
```
🪙 Bitcoin (BTC)
Rank: #1
Price: $95539.00

📊 Price Changes
24h: -1.28%
7d: -7.36%
30d: -3.85%

📈 Market Data
Market Cap: $1.89T
24h Volume: $40.93B
Circulating Supply: 19.80M BTC
Max Supply: 21.00M BTC

🏆 All Time High/Low
ATH: $108135.00 (2024-12-17)
ATL: $67.81 (2013-07-06)
```

### Interactive Charts

```bash
# Display line chart (default)
crypto bitcoin --graph

# Show candlestick chart
crypto bitcoin --graph --candles

# Customize time interval (default: 7d)
crypto bitcoin --graph --interval 30d

# Available intervals: 1d, 7d, 14d, 30d, 90d, 180d, 1y, max
```

### Portfolio Management

```bash
# Add a buy transaction
crypto portfolio add bitcoin 0.5 50000 buy  # <coin> <amount> <price> <buy/sell>

# Add a sell transaction
crypto portfolio add bitcoin 0.1 55000 sell

# View portfolio holdings with current values
crypto portfolio list

# Check transaction history
crypto portfolio history

# Remove a specific coin from portfolio
crypto portfolio remove bitcoin  # Will remove the coin and all its transactions

# Clear entire portfolio
crypto portfolio clear  # Will remove all coins and transactions

# Change currency for portfolio valuation
crypto portfolio list --currency eur
```

Example output for portfolio list:
```
💼 Portfolio Holdings

       COIN      |   AMOUNT    |  PRICE  |  VALUE   | 24H CHANGE  
-----------------+-------------+---------+----------+-------------
  Bitcoin (BTC)  | 1.05        | $95.33K | $100.10K | -1.70%
  Ethereum (ETH) | 2000        | $3.27K  | $6.55M   | -2.07%
-----------------+-------------+---------+----------+-------------
   TOTAL VALUE   |                          $6.65M  |
-----------------+-------------+---------+----------+-------------
```

Example output for portfolio history:
```
📜 Transaction History

       DATE        |  TYPE  |     COIN      |  AMOUNT   |  PRICE   
-------------------+--------+---------------+-----------+----------
  2024-01-22 10:30 | BUY    | Bitcoin (BTC) | 0.50000   | $95000.00
  2024-01-22 15:45 | SELL   | Bitcoin (BTC) | 0.10000   | $96500.00
  2024-01-23 09:15 | BUY    | Ethereum (ETH)| 2000.00000| $3270.00
```

Example output for portfolio remove:
```
Are you sure you want to remove BITCOIN (Amount: 0.500000) from your portfolio? (y/N): y

💼 BITCOIN removed from portfolio successfully
```

Example output for portfolio clear:
```
Are you sure you want to clear your entire portfolio? (y/N): y

💼 Portfolio cleared successfully
```

### Price Alerts

```bash
# Set price alert for when Bitcoin goes above $50,000
crypto alert add bitcoin 50000 above  # <coin> <price> <above/below>

# Set price alert for when Bitcoin goes below $45,000
crypto alert add bitcoin 45000 below

# List active alerts
crypto alert list

# Remove specific alert
crypto alert remove bitcoin
```

Example output:
```
🔔 Active Price Alerts

    COIN   | CONDITION | TARGET PRICE |    CREATED AT     
-----------+-----------+--------------+-------------------
  ETHEREUM | below     | $3000.00     | 2024-12-22 20:48
  BITCOIN  | above     | $100000.00   | 2024-12-22 21:00
```

## API Rate Limits

This tool uses CoinGecko's public API. Please be mindful of rate limits when making frequent requests.

## Contributing

Contributions are welcome! Feel free to submit issues and pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Data provided by [CoinGecko](https://www.coingecko.com/)
- Built with [Go](https://golang.org/)
