# crypto

<p><a href="https://go.dev" target="_blank"><img src="https://img.shields.io/badge/Go-1.16+-00ADD8?style=for-the-badge&logo=go" alt="go version" /></a>&nbsp;<img src="https://img.shields.io/badge/license-apache_2.0-red?style=for-the-badge&logo=none" alt="license" />
</p>

<p>This CLI Tool instantly retrieve prices and other information of cryptocurrencies.</p>

![Cryptocurrency-CLI](https://github.com/mrcnserkan/crypto/blob/master/crypto.png)

## Installation

First, [download](https://golang.org/dl/) and install **Go**.

Installation is done by using the [`go install`](https://golang.org/cmd/go/#hdr-Compile_and_install_packages_and_dependencies) command and rename installed binary in `$GOPATH/bin`:

```bash
go install github.com/mrcnserkan/crypto@latest
```

## Usage

### Default Command

```bash
# `crypto -h` for help

crypto
```

### Page Through The Results

```bash
# Default value 1

crypto --page 2
```

### Total Results Per Page

```bash
# Default value 10

crypto --per-page 5
```

### The Target Currency of Market Data

```bash
# Default value usd

crypto --currency try
```

### Search

Search for coins listed on CoinGecko ordered by largest Market Cap first

```bash
crypto --search solana
```

### Coin Detail

```bash
# You can learn coin-id with search flag

crypto ethereum
```

###

![Cryptocurrency-CLI](https://github.com/mrcnserkan/crypto/blob/master/crypto.gif)
