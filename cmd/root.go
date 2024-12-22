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
	"github.com/guptarohit/asciigraph"
	"github.com/mrcnserkan/crypto/models"
	"github.com/mrcnserkan/crypto/service"
	"github.com/mrcnserkan/crypto/utils"
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
	configDir    string
	rootCmd      *cobra.Command
)

const Version = "v1.2.2"

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
     crypto bitcoin --graph         # Show line chart
     crypto bitcoin --graph --candles    # Show candlestick chart
     crypto bitcoin --graph --interval 30d   # Show 30-day chart

  4. Manage portfolio:
     crypto portfolio add bitcoin 0.5 50000 buy   # Add transaction
     crypto portfolio list                        # View holdings
     crypto portfolio history                     # View history

  5. Set price alerts:
     crypto alert add bitcoin 50000 above   # Alert when price goes above
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

			currency, _ := cmd.Flags().GetString("currency")
			if len(args) > 0 {
				showGraph, _ := cmd.Flags().GetBool("graph")

				if showGraph {
					displayPriceGraph(args[0], currency)
				} else {
					coinDetail, err := coinGecko.GetCoinDetail(args[0])
					if err != nil || coinDetail.ID == "" {
						fmt.Printf("Error: Could not find coin with ID '%s'\n", args[0])
						os.Exit(1)
					}
					PrintCoinDetail(args[0], currency)
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

	// Add subcommands
	rootCmd.AddCommand(alertCmd)
	rootCmd.AddCommand(portfolioCmd)

	// Initialize configuration
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	configDir = filepath.Join(homeDir, ".crypto")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	portfolio = models.NewPortfolio(filepath.Join(configDir, "portfolio.json"))
	alertManager = models.NewAlertManager(configDir)
	alertChecker = service.NewAlertChecker(alertManager)
	coinGecko = service.NewCoinGecko()

	// Load existing portfolio and alerts
	if err := portfolio.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Println("Error loading portfolio:", err)
	}
	if err := alertManager.Load(); err != nil {
		fmt.Println("Error loading alerts:", err)
	}

	// Start alert checker
	alertChecker.Start()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nShutting down...")
		alertChecker.Stop()
		os.Exit(0)
	}()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func displayPriceGraph(coinID, currency string) {
	currencySymbol := service.DEFAULT_CURRENCY_SYMBOL
	if currency != service.DEFAULT_CURRENCY {
		currencySymbol = strings.ToUpper(currency)
	}

	// Get interval and chart type from flags
	interval, _ := rootCmd.Flags().GetString("interval")
	showCandles, _ := rootCmd.Flags().GetBool("candles")

	var priceValues []float64
	var timestamps []time.Time
	var ohlcData []models.OHLC

	if showCandles {
		// Get OHLC data for candle chart
		data, err := coinGecko.GetCoinOHLC(coinID, currency, interval)
		if err != nil {
			fmt.Printf("Error fetching OHLC data: %v\n", err)
			os.Exit(1)
		}
		ohlcData = data
		// Extract close prices for price range
		priceValues = make([]float64, len(data))
		for i, candle := range data {
			priceValues[i] = candle.Close
			timestamps = append(timestamps, time.Unix(candle.Time/1000, 0))
		}
	} else {
		// Get price history for line chart
		prices, err := coinGecko.GetCoinPriceHistory(coinID, currency, interval)
		if err != nil {
			fmt.Printf("Error fetching price history: %v\n", err)
			os.Exit(1)
		}

		// Extract price values and timestamps
		priceValues = make([]float64, len(prices))
		for i, price := range prices {
			priceValues[i] = price[1]
			timestamps = append(timestamps, time.Unix(int64(price[0])/1000, 0))
		}
	}

	// Color definitions
	titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
	captionColor := color.New(color.FgHiBlue).SprintFunc()
	labelColor := color.New(color.FgHiBlue).SprintFunc()
	valueColor := color.New(color.FgHiWhite).SprintFunc()

	// Title
	fmt.Printf("\n%s %s\n\n", titleColor("📈"), titleColor(fmt.Sprintf("%s Price Chart (%s)", strings.ToUpper(coinID), interval)))

	// Price range information
	minPrice, maxPrice := utils.MinMax(priceValues)
	priceRange := maxPrice - minPrice
	fmt.Printf("%s %s%s - %s%s (Δ %s%s)\n",
		labelColor("Price Range:"),
		currencySymbol, valueColor(utils.FormatCurrency(minPrice)),
		currencySymbol, valueColor(utils.FormatCurrency(maxPrice)),
		currencySymbol, valueColor(utils.FormatCurrency(priceRange)))

	// Time range information
	if len(timestamps) > 0 {
		startTime := timestamps[0]
		endTime := timestamps[len(timestamps)-1]
		fmt.Printf("%s %s - %s\n\n",
			labelColor("Time Range:"),
			valueColor(startTime.Format("2006-01-02 15:04")),
			valueColor(endTime.Format("2006-01-02 15:04")))
	}

	if showCandles {
		// Draw candle chart
		candleChart := utils.GenerateCandleChart(ohlcData)
		fmt.Println(candleChart)
		fmt.Println(captionColor(fmt.Sprintf("Data source from coingecko.com at %s", utils.GetCurrentTime())))
	} else {
		// Prepare Y-axis labels
		priceStep := priceRange / 4
		fmt.Printf("%s%s ┤\n", currencySymbol, utils.FormatCurrency(maxPrice))
		for i := 3; i >= 0; i-- {
			price := minPrice + (float64(i) * priceStep)
			fmt.Printf("%s%s ┤\n", currencySymbol, utils.FormatCurrency(price))
		}

		// Draw line chart
		graph := asciigraph.Plot(priceValues,
			asciigraph.Height(20),   // Increase height
			asciigraph.Width(100),   // Fixed width
			asciigraph.Precision(2), // 2 decimal precision
		)
		fmt.Println(graph)
		fmt.Println(captionColor(fmt.Sprintf("\nData source from coingecko.com at %s", utils.GetCurrentTime())))
	}
}

func PrintList(page string, perPage string, currency string) {
	currencySymbol := service.DEFAULT_CURRENCY_SYMBOL
	if currency != service.DEFAULT_CURRENCY {
		currencySymbol = strings.ToUpper(currency)
	}

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

	pageNum := utils.StringToInt(page)
	perPageNum := utils.StringToInt(perPage)
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

func PrintCoinDetail(coinName string, currency string) {
	currencySymbol := service.DEFAULT_CURRENCY_SYMBOL
	if currency == "" {
		currency = service.DEFAULT_CURRENCY
	}
	if currency != service.DEFAULT_CURRENCY {
		currencySymbol = strings.ToUpper(currency)
	}

	coinDetail, err := coinGecko.GetCoinDetail(coinName)
	if err != nil || coinDetail.ID == "" {
		fmt.Printf("Error: Could not find coin with ID '%s'\n", coinName)
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
	fmt.Printf("%s %s%s\n", labelColor("Price:"), currencySymbol, valueColor(fmt.Sprintf("%.2f", coinDetail.MarketData.CurrentPrice[currency])))

	// Price changes
	change24h := coinDetail.MarketData.PriceChangePercentage24HInCurrency[currency]
	change7d := coinDetail.MarketData.PriceChangePercentage7DInCurrency[currency]
	change30d := coinDetail.MarketData.PriceChangePercentage30DInCurrency[currency]

	fmt.Printf("\n%s\n", titleColor("📊 Price Changes"))
	fmt.Printf("%s %s\n", labelColor("24h:"), formatPriceChange(change24h, green, red))
	fmt.Printf("%s %s\n", labelColor("7d:"), formatPriceChange(change7d, green, red))
	fmt.Printf("%s %s\n", labelColor("30d:"), formatPriceChange(change30d, green, red))

	// Market data
	fmt.Printf("\n%s\n", titleColor("📈 Market Data"))
	fmt.Printf("%s %s%s\n", labelColor("Market Cap:"), currencySymbol, valueColor(utils.FormatCurrency(float64(coinDetail.MarketData.MarketCap[currency]))))
	fmt.Printf("%s %s%s\n", labelColor("24h Volume:"), currencySymbol, valueColor(utils.FormatCurrency(coinDetail.MarketData.TotalVolume[currency])))
	fmt.Printf("%s %s %s\n", labelColor("Circulating Supply:"), valueColor(utils.FormatCurrency(coinDetail.MarketData.CirculatingSupply)), strings.ToUpper(coinDetail.Symbol))
	if coinDetail.MarketData.MaxSupply > 0 {
		fmt.Printf("%s %s %s\n", labelColor("Max Supply:"), valueColor(utils.FormatCurrency(coinDetail.MarketData.MaxSupply)), strings.ToUpper(coinDetail.Symbol))
	}

	// ATH/ATL information
	fmt.Printf("\n%s\n", titleColor("🏆 All Time High/Low"))
	fmt.Printf("%s %s%s (%s)\n", labelColor("ATH:"), currencySymbol, valueColor(fmt.Sprintf("%.2f", coinDetail.MarketData.Ath[currency])), valueColor(coinDetail.MarketData.AthDate[currency][:10]))
	fmt.Printf("%s %s%s (%s)\n", labelColor("ATL:"), currencySymbol, valueColor(fmt.Sprintf("%.2f", coinDetail.MarketData.Atl[currency])), valueColor(coinDetail.MarketData.AtlDate[currency][:10]))
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
