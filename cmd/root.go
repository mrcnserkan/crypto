/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

package cmd

import (
	"fmt"
	"os"

	"github.com/mrcnserkan/crypto/service"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "crypto [coin-id]",
	Short: "A simple cryptocurrency price checker in your console.",
	Long: `Check the prices of cryptocurrencies, 24h and 7d percentage price changes, market capacities and all-time highs.
For a cryptocurrency detail e.g. "crypto bitcoin"
For search e.g. "crypto --search bitcoin"
For page by page view e.g. "crypto --page 1 --per-page 20"
For a custom currency e.g. "crypto --currency eur" or "crypto ethereum --currency try"`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		currency, _ := cmd.Flags().GetString("currency")
		if len(args) > 0 {
			coinName := args[0]
			PrintCoinDetail(coinName, currency)
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
		fmt.Println()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("page", service.DEFAULT_PAGE, "Page through results")
	rootCmd.PersistentFlags().String("per-page", service.PER_PAGE, "Total results per page")
	rootCmd.PersistentFlags().String("currency", service.DEFAULT_CURRENCY, "The target currency of market data")
	rootCmd.PersistentFlags().String(
		"search",
		"",
		"Search for coins listed on CoinGecko ordered by largest Market Cap first",
	)
}
