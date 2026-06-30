package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/mrcnserkan/crypto/v2/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var watchlistCmd = &cobra.Command{
	Use:   "watchlist",
	Short: "Manage coin watchlist",
	Long: `Add, remove, and view coins on your personal watchlist.

EXAMPLES:
  crypto watchlist add bitcoin
  crypto watchlist remove ethereum
  crypto watchlist list`,
}

var watchlistAddCmd = &cobra.Command{
	Use:   "add [coin-id]",
	Short: "Add coin to watchlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		coinID := utils.NormalizeCoinID(args[0])
		coin, err := coinGecko.GetCoinDetail(coinID)
		if err != nil || coin.ID == "" {
			fmt.Printf("Error: Could not find coin with ID '%s'\n", args[0])
			os.Exit(1)
		}
		if err := watchlist.Add(coinID); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s %s added to watchlist\n", titleColor("⭐"), titleColor(coin.Name))
	},
}

var watchlistRemoveCmd = &cobra.Command{
	Use:   "remove [coin-id]",
	Short: "Remove coin from watchlist",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		coinID := utils.NormalizeCoinID(args[0])
		if err := watchlist.Remove(coinID); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s %s removed from watchlist\n", titleColor("⭐"), titleColor(strings.ToUpper(coinID)))
	},
}

var watchlistListCmd = &cobra.Command{
	Use:   "list",
	Short: "List watchlist coins",
	Run: func(cmd *cobra.Command, args []string) {
		if len(watchlist.CoinIDs) == 0 {
			fmt.Println("\nWatchlist is empty")
			return
		}

		currency := getCurrencyFlag(cmd)
		currencySymbol := utils.CurrencySymbol(currency)

		coins, err := coinGecko.GetMarketsByIDs(currency, watchlist.CoinIDs)
		if err != nil {
			fmt.Printf("Error fetching market data: %v\n", err)
			os.Exit(1)
		}

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s %s\n\n", titleColor("⭐"), titleColor("Watchlist"))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Coin", "Price", "24h", "7d"})
		table.SetBorder(false)

		for _, coin := range coins {
			table.Rich([]string{
				fmt.Sprintf("%s (%s)", coin.Name, strings.ToUpper(coin.Symbol)),
				fmt.Sprintf("%s%s", currencySymbol, utils.FormatCurrency(coin.CurrentPrice)),
				fmt.Sprintf("%.1f%%", coin.PriceChangePercentage24h),
				fmt.Sprintf("%.1f%%", coin.PriceChangePercentage7DInCurrency),
			}, []tablewriter.Colors{
				{tablewriter.FgHiWhiteColor},
				{tablewriter.FgHiWhiteColor},
				utils.GetCellColorFromPriceChange(coin.PriceChangePercentage24h),
				utils.GetCellColorFromPriceChange(coin.PriceChangePercentage7DInCurrency),
			})
		}
		table.Render()
	},
}

func init() {
	watchlistCmd.AddCommand(watchlistAddCmd)
	watchlistCmd.AddCommand(watchlistRemoveCmd)
	watchlistCmd.AddCommand(watchlistListCmd)
	rootCmd.AddCommand(watchlistCmd)
}
