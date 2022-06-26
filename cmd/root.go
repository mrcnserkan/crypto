/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

package cmd

import (
	"os"

	"github.com/mrcnserkan/crypto/service"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cryptocurrency-cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
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
