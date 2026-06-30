/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/mrcnserkan/crypto/v2/models"
	"github.com/mrcnserkan/crypto/v2/service"
	"github.com/mrcnserkan/crypto/v2/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	portfolio    *models.Portfolio
	alertManager *models.AlertManager
	alertChecker *service.AlertChecker
	coinGecko    *service.CoinGecko
	configStore  *models.ConfigStore
	watchlist    *models.Watchlist
	daemonState  *models.DaemonState
	configDir    string
	rootCmd      *cobra.Command
)

const Version = "v2.0.1"

func init() {
	rootCmd = &cobra.Command{
		Use:   "crypto [coin-id]",
		Short: "A powerful cryptocurrency tracking CLI tool",
		Long: `Crypto CLI - Real-time cryptocurrency tracking and portfolio management

BASIC USAGE:
  crypto                    List top cryptocurrencies
  crypto bitcoin           Show detailed information for Bitcoin
  crypto --search solana   Search for coins by name or symbol

MAIN FEATURES:
  • Real-time price tracking and market data
  • Portfolio management with transaction history
  • Price alerts with notifications
  • Interactive price charts
  • Multi-currency support (USD, EUR, TRY, etc.)

EXAMPLES:
  1. View top cryptocurrencies:
     crypto                          # Default: page 1, 10 results
     crypto --page 2                 # View next page
     crypto --per-page 20           # Show 20 results per page
     crypto --currency eur          # Show prices in EUR

  2. Get detailed coin information:
     crypto bitcoin                 # Show Bitcoin details
     crypto ethereum               # Show Ethereum details
     crypto --search "solana"      # Search for coins

  3. View price charts:
     crypto bitcoin --graph                    # Line chart (7 days)
     crypto bitcoin --graph --candles          # Candlestick chart
     crypto bitcoin --graph --interval 30d   # 30-day chart
     crypto bitcoin --graph --from 2026-06-01 --to 2026-06-30
     crypto bitcoin --graph --width 100 --height 24

  4. Manage portfolio:
     crypto portfolio add bitcoin 0.5 50000 buy   # Add transaction
     crypto portfolio list                        # View holdings
     crypto portfolio history                     # View history

  5. Set price alerts:
     crypto alert add bitcoin 50000 above   # Alert when price goes above
     crypto alert watch                     # Monitor alerts (foreground)
     crypto alert start                     # Monitor alerts (background)
     crypto alert list                      # View active alerts

Use "crypto [command] --help" for more information about a command.`,
		Version: Version,
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Show version and exit if --version flag is used
			if v, _ := cmd.Flags().GetBool("version"); v {
				fmt.Printf("Crypto CLI %s\n", Version)
				os.Exit(0)
			}

			currency := getCurrencyFlag(cmd)
			if len(args) > 0 {
				showGraph, _ := cmd.Flags().GetBool("graph")

				if showGraph {
					displayPriceGraph(utils.NormalizeCoinID(args[0]), currency)
				} else {
					coinDetail, err := coinGecko.GetCoinDetail(utils.NormalizeCoinID(args[0]))
					if err != nil || coinDetail.ID == "" {
						fmt.Printf("Error: Could not find coin with ID '%s'\n", args[0])
						os.Exit(1)
					}
					PrintCoinDetail(coinDetail, currency)
				}
			} else {
				search, _ := cmd.Flags().GetString("search")
				if search != "" {
					PrintSearchResult(search)
				} else {
					page, _ := cmd.Flags().GetString("page")
					perPage, _ := cmd.Flags().GetString("per-page")
					PrintList(page, perPage, currency)
				}
			}
		},
	}

	// Root command flags
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	rootCmd.PersistentFlags().String("currency", service.DEFAULT_CURRENCY, "Currency for price display (e.g., usd, eur, gbp)")
	rootCmd.PersistentFlags().String("page", service.DEFAULT_PAGE, "Page number for paginated results")
	rootCmd.PersistentFlags().String("per-page", service.PER_PAGE, "Number of results per page (max: 250)")
	rootCmd.PersistentFlags().String("search", "", "Search for cryptocurrencies by name or symbol")
	rootCmd.PersistentFlags().Bool("graph", false, "Display price chart for the specified coin")
	rootCmd.PersistentFlags().String("interval", "7d", "Chart time interval (1d, 7d, 14d, 30d, 90d, 180d, 1y, max)")
	rootCmd.PersistentFlags().Bool("candles", false, "Display candlestick chart instead of line chart")
	rootCmd.PersistentFlags().String("from", "", "Chart start date (YYYY-MM-DD or YYYY-MM-DD HH:MM)")
	rootCmd.PersistentFlags().String("to", "", "Chart end date (YYYY-MM-DD or YYYY-MM-DD HH:MM)")
	rootCmd.PersistentFlags().Int("width", 80, "Chart width in characters")
	rootCmd.PersistentFlags().Int("height", 20, "Chart height in characters")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored terminal output")

	// Add subcommands
	rootCmd.AddCommand(alertCmd)
	rootCmd.AddCommand(portfolioCmd)
	rootCmd.AddCommand(newCompletionCmd())

	// Initialize configuration
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	configDir = filepath.Join(homeDir, ".crypto")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	portfolio = models.NewPortfolio(filepath.Join(configDir, "portfolio.json"))
	alertManager = models.NewAlertManager(configDir)
	configStore = models.NewConfigStore(configDir)
	watchlist = models.NewWatchlist(configDir)
	daemonState = models.NewDaemonState(configDir)
	alertChecker = service.NewAlertChecker(alertManager)
	coinGecko = service.NewCoinGecko()

	if err := configStore.Load(); err != nil {
		fmt.Println("Error loading config:", err)
	}
	if err := watchlist.Load(); err != nil {
		fmt.Println("Error loading watchlist:", err)
	}

	// Load existing portfolio and alerts
	if err := portfolio.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Println("Error loading portfolio:", err)
	}
	if err := alertManager.Load(); err != nil {
		fmt.Println("Error loading alerts:", err)
	}
}

func Execute() {
	disableColorsIfNeeded()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nShutting down...")
		alertChecker.Stop()
		os.Exit(0)
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func displayPriceGraph(coinID, currency string) {
	currency = utils.NormalizeCurrency(currency)
	currencySymbol := utils.CurrencySymbol(currency)

	interval, _ := rootCmd.Flags().GetString("interval")
	showCandles, _ := rootCmd.Flags().GetBool("candles")
	fromStr, _ := rootCmd.Flags().GetString("from")
	toStr, _ := rootCmd.Flags().GetString("to")
	chartWidth, _ := rootCmd.Flags().GetInt("width")
	chartHeight, _ := rootCmd.Flags().GetInt("height")
	if configStore != nil {
		if chartWidth == 80 {
			chartWidth = configStore.ChartWidthOrDefault(80)
		}
		if chartHeight == 20 {
			chartHeight = configStore.ChartHeightOrDefault(20)
		}
	}

	var fromDate, toDate *time.Time
	if fromStr != "" {
		t, err := utils.ParseChartDate(fromStr)
		if err != nil {
			fmt.Printf("Error: invalid --from date: %v\n", err)
			os.Exit(1)
		}
		fromDate = &t
	}
	if toStr != "" {
		t, err := utils.ParseChartDate(toStr)
		if err != nil {
			fmt.Printf("Error: invalid --to date: %v\n", err)
			os.Exit(1)
		}
		// Include the full end day when only a date is given
		if !strings.Contains(toStr, ":") {
			endOfDay := t.Add(24*time.Hour - time.Second)
			toDate = &endOfDay
		} else {
			toDate = &t
		}
	}

	// When custom range is set, pick API interval that covers the span
	apiInterval := service.SelectInterval(interval)
	if fromDate != nil || toDate != nil {
		rangeFrom := time.Now().AddDate(0, 0, -7)
		rangeTo := time.Now()
		if fromDate != nil {
			rangeFrom = *fromDate
		}
		if toDate != nil {
			rangeTo = *toDate
		}
		apiInterval = service.SelectIntervalForRange(rangeFrom, rangeTo)
		interval = apiInterval.Name
	}

	chartCfg := utils.ChartConfig{
		Width:          chartWidth,
		Height:         chartHeight,
		CurrencySymbol: currencySymbol,
		YTickCount:     5,
		XTickCount:     5,
	}

	var series []utils.SeriesPoint
	var ohlcData []models.OHLC

	if showCandles {
		data, err := coinGecko.GetCoinOHLC(coinID, currency, apiInterval.Name)
		if err != nil {
			fmt.Printf("Error fetching OHLC data: %v\n", err)
			os.Exit(1)
		}
		ohlcData = utils.FilterOHLCByDateRange(data, fromDate, toDate)
		for _, candle := range ohlcData {
			series = append(series, utils.SeriesPoint{
				Time:  time.Unix(candle.Time/1000, 0),
				Value: candle.Close,
			})
		}
	} else {
		prices, err := coinGecko.GetCoinPriceHistory(coinID, currency, apiInterval.Name)
		if err != nil {
			fmt.Printf("Error fetching price history: %v\n", err)
			os.Exit(1)
		}
		for _, price := range prices {
			series = append(series, utils.SeriesPoint{
				Time:  time.Unix(int64(price[0])/1000, 0),
				Value: price[1],
			})
		}
		series = utils.FilterSeriesByDateRange(series, fromDate, toDate)
	}

	if len(series) == 0 {
		fmt.Println("Error: No price data available for the selected interval or date range")
		os.Exit(1)
	}

	chartType := "Line"
	if showCandles {
		chartType = "Candlestick"
	}

	titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	captionColor := color.New(color.FgHiBlue).SprintFunc()
	labelColor := color.New(color.FgHiBlue).SprintFunc()
	valueColor := color.New(color.FgHiWhite).SprintFunc()
	statsColor := color.New(color.FgHiYellow).SprintFunc()

	fmt.Printf("\n%s %s\n\n", titleColor("📈"), titleColor(fmt.Sprintf("%s %s Chart (%s)",
		strings.ToUpper(coinID), chartType, interval)))

	stats := utils.ComputePeriodStats(series)
	fmt.Println(statsColor(utils.FormatPeriodStatsLine(stats, currencySymbol)))

	minPrice, maxPrice := series[0].Value, series[0].Value
	for _, p := range series[1:] {
		if p.Value < minPrice {
			minPrice = p.Value
		}
		if p.Value > maxPrice {
			maxPrice = p.Value
		}
	}
	priceRange := maxPrice - minPrice
	fmt.Printf("%s %s%s - %s%s (Δ %s%s)\n",
		labelColor("Price Range:"),
		currencySymbol, valueColor(utils.FormatCurrency(minPrice)),
		currencySymbol, valueColor(utils.FormatCurrency(maxPrice)),
		currencySymbol, valueColor(utils.FormatCurrency(priceRange)))

	startTime := series[0].Time
	endTime := series[len(series)-1].Time
	fmt.Printf("%s %s - %s\n\n",
		labelColor("Time Range:"),
		valueColor(startTime.Format("2006-01-02 15:04")),
		valueColor(endTime.Format("2006-01-02 15:04")))

	if showCandles {
		fmt.Print(utils.RenderCandleChart(ohlcData, chartCfg))
	} else {
		fmt.Print(utils.RenderLineChart(series, chartCfg))
	}

	legend := "█/░ = Bullish/Bearish candles"
	if !showCandles {
		legend = "● = price point"
	}
	fmt.Println(captionColor(fmt.Sprintf("%s | Data source: coingecko.com at %s", legend, utils.GetCurrentTime())))
}

func PrintList(page string, perPage string, currency string) {
	pageNum, err := utils.ParsePage(page)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	perPageNum, err := utils.ParsePerPage(perPage)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	currency = utils.NormalizeCurrency(currency)
	currencySymbol := utils.CurrencySymbol(currency)

	printer := message.NewPrinter(language.English)
	table := tablewriter.NewWriter(os.Stdout)

	// Color definitions
	titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()

	fmt.Printf("\n%s %s\n\n", titleColor("🏆"), titleColor("Top Cryptocurrencies by Market Cap"))

	table.SetHeader([]string{"#", "Coin", "", "Price", "24h", "7d", "Market Cap", "ATH"})
	table.SetBorder(false)
	table.SetRowLine(true)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
	)

	coins, err := coinGecko.GetMarkets(currency, perPageNum, pageNum)
	if err != nil {
		fmt.Printf("Error fetching market data: %v\n", err)
		os.Exit(1)
	}

	for _, coin := range coins {
		table.Rich([]string{
			fmt.Sprintf("%d", coin.MarketCapRank),
			coin.Name,
			strings.ToUpper(coin.Symbol),
			printer.Sprintf("%s%s", currencySymbol, utils.FormatCurrency(coin.CurrentPrice)),
			fmt.Sprintf("%.1f%%", coin.PriceChangePercentage24h),
			fmt.Sprintf("%.1f%%", coin.PriceChangePercentage7DInCurrency),
			printer.Sprintf("%s%s", currencySymbol, utils.FormatCurrency(coin.MarketCap)),
			printer.Sprintf("%s%s", currencySymbol, utils.FormatCurrency(coin.Ath)),
		}, []tablewriter.Colors{
			{tablewriter.FgHiWhiteColor},
			{tablewriter.FgHiWhiteColor},
			{tablewriter.FgHiWhiteColor},
			{tablewriter.Bold, tablewriter.FgHiWhiteColor},
			utils.GetCellColorFromPriceChange(coin.PriceChangePercentage24h),
			utils.GetCellColorFromPriceChange(coin.PriceChangePercentage7DInCurrency),
			{tablewriter.FgHiWhiteColor},
			{tablewriter.FgHiWhiteColor},
		})
	}

	table.SetCaption(true, utils.GetCaption())
	table.Render()
}

func PrintCoinDetail(coinDetail models.CoinDetail, currency string) {
	currency = utils.NormalizeCurrency(currency)
	currencySymbol := utils.CurrencySymbol(currency)

	currentPrice, err := utils.PriceFromCurrencyMap(coinDetail.MarketData.CurrentPrice, currency)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	marketCap, err := utils.Int64FromCurrencyMap(coinDetail.MarketData.MarketCap, currency)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	totalVolume, err := utils.FloatFromCurrencyMap(coinDetail.MarketData.TotalVolume, currency)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	ath, err := utils.FloatFromCurrencyMap(coinDetail.MarketData.Ath, currency)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	atl, err := utils.FloatFromCurrencyMap(coinDetail.MarketData.Atl, currency)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	change24h, err := utils.FloatFromCurrencyMap(coinDetail.MarketData.PriceChangePercentage24HInCurrency, currency)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	change7d, err := utils.FloatFromCurrencyMap(coinDetail.MarketData.PriceChangePercentage7DInCurrency, currency)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	change30d, err := utils.FloatFromCurrencyMap(coinDetail.MarketData.PriceChangePercentage30DInCurrency, currency)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	athDate, athDateOK := coinDetail.MarketData.AthDate[currency]
	atlDate, atlDateOK := coinDetail.MarketData.AtlDate[currency]
	if !athDateOK || !atlDateOK {
		fmt.Printf("Error: unsupported or invalid currency: %s\n", strings.ToUpper(currency))
		os.Exit(1)
	}

	// Color definitions
	titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	labelColor := color.New(color.FgHiBlue).SprintFunc()
	valueColor := color.New(color.FgHiWhite).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Main information
	fmt.Printf("\n%s %s\n", titleColor("🪙"), titleColor(fmt.Sprintf("%s (%s)", coinDetail.Name, strings.ToUpper(coinDetail.Symbol))))
	fmt.Printf("%s %s\n", labelColor("Rank:"), valueColor(fmt.Sprintf("#%d", coinDetail.MarketCapRank)))
	fmt.Printf("%s %s%s\n", labelColor("Price:"), currencySymbol, valueColor(fmt.Sprintf("%.2f", currentPrice)))

	// Price changes
	fmt.Printf("\n%s\n", titleColor("📊 Price Changes"))
	fmt.Printf("%s %s\n", labelColor("24h:"), formatPriceChange(change24h, green, red))
	fmt.Printf("%s %s\n", labelColor("7d:"), formatPriceChange(change7d, green, red))
	fmt.Printf("%s %s\n", labelColor("30d:"), formatPriceChange(change30d, green, red))

	// Market data
	fmt.Printf("\n%s\n", titleColor("📈 Market Data"))
	fmt.Printf("%s %s%s\n", labelColor("Market Cap:"), currencySymbol, valueColor(utils.FormatCurrency(float64(marketCap))))
	fmt.Printf("%s %s%s\n", labelColor("24h Volume:"), currencySymbol, valueColor(utils.FormatCurrency(totalVolume)))
	fmt.Printf("%s %s %s\n", labelColor("Circulating Supply:"), valueColor(utils.FormatCurrency(coinDetail.MarketData.CirculatingSupply)), strings.ToUpper(coinDetail.Symbol))
	if coinDetail.MarketData.MaxSupply > 0 {
		fmt.Printf("%s %s %s\n", labelColor("Max Supply:"), valueColor(utils.FormatCurrency(coinDetail.MarketData.MaxSupply)), strings.ToUpper(coinDetail.Symbol))
	}

	// ATH/ATL information
	fmt.Printf("\n%s\n", titleColor("🏆 All Time High/Low"))
	fmt.Printf("%s %s%s (%s)\n", labelColor("ATH:"), currencySymbol, valueColor(fmt.Sprintf("%.2f", ath)), valueColor(utils.FormatISODate(athDate)))
	fmt.Printf("%s %s%s (%s)\n", labelColor("ATL:"), currencySymbol, valueColor(fmt.Sprintf("%.2f", atl)), valueColor(utils.FormatISODate(atlDate)))
}

func formatPriceChange(change float64, green, red func(a ...interface{}) string) string {
	if change >= 0 {
		return green(fmt.Sprintf("+%.2f%%", change))
	}
	return red(fmt.Sprintf("%.2f%%", change))
}

func PrintSearchResult(query string) {
	searchResult, err := coinGecko.SearchCoins(query)
	if err != nil {
		fmt.Printf("Error searching coins: %v\n", err)
		os.Exit(1)
	}

	// Color definitions
	titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()

	fmt.Printf("\n%s %s\n\n", titleColor("🔍"), titleColor(fmt.Sprintf("Search Results for '%s'", query)))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Coin", "Symbol", "Market Cap Rank"})
	table.SetBorder(false)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgHiBlueColor},
	)

	for i, coin := range searchResult.Coins {
		rank := "-"
		if coin.MarketCapRank > 0 {
			rank = fmt.Sprintf("%d", coin.MarketCapRank)
		}

		table.Rich([]string{
			fmt.Sprintf("%d", i+1),
			coin.Name,
			strings.ToUpper(coin.Symbol),
			rank,
		}, []tablewriter.Colors{
			{tablewriter.FgHiWhiteColor},
			{tablewriter.FgHiWhiteColor},
			{tablewriter.FgHiWhiteColor},
			{tablewriter.FgHiWhiteColor},
		})
	}

	table.Render()
}
