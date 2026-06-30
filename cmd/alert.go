/*
Copyright © 2024 Serkan MERCAN <serkanmercan@email.com>
*/

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/mrcnserkan/crypto/v2/models"
	"github.com/mrcnserkan/crypto/v2/service"
	"github.com/mrcnserkan/crypto/v2/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var alertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Price alert management",
	Long: `Set and manage price alerts for cryptocurrencies.

AVAILABLE COMMANDS:
  add     Set a new price alert
  list    View active alerts
  remove  Remove specific alerts

EXAMPLES:
  1. Set price alerts:
     crypto alert add bitcoin 50000 above    # Alert when BTC goes above $50,000
     crypto alert add ethereum 2000 below    # Alert when ETH goes below $2,000

  2. View alerts:
     crypto alert list                       # Show all active alerts

  3. Remove alerts:
     crypto alert remove bitcoin             # Remove all alerts for a coin
     crypto alert remove bitcoin 50000 above # Remove a specific alert

  4. Monitor alerts:
     crypto alert watch                      # Foreground monitoring
     crypto alert start                      # Background daemon
     crypto alert stop                       # Stop daemon
     crypto alert status                     # Daemon status

NOTE: Alerts are NOT checked automatically. Run 'crypto alert watch' or 'crypto alert start' after adding alerts.`,
}

var alertAddCmd = &cobra.Command{
	Use:   "add [coin-id] [price] [above/below]",
	Short: "Add price alert",
	Long: `Set a new price alert for a cryptocurrency.

ARGUMENTS:
  coin-id  ID of the cryptocurrency (e.g., bitcoin)
  price    Target price for the alert
  type     Alert type: 'above' or 'below'

EXAMPLES:
  crypto alert add bitcoin 50000 above    # Alert when BTC goes above $50,000
  crypto alert add ethereum 2000 below    # Alert when ETH goes below $2,000`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		coinID := utils.NormalizeCoinID(args[0])
		price, err := strconv.ParseFloat(args[1], 64)
		if err != nil || price <= 0 {
			fmt.Println("Error: Invalid price value (must be greater than zero)")
			os.Exit(1)
		}
		condition := strings.ToLower(args[2])
		if condition != "above" && condition != "below" {
			fmt.Println("Error: Condition must be 'above' or 'below'")
			os.Exit(1)
		}

		currency, _ := rootCmd.PersistentFlags().GetString("currency")
		currency = utils.NormalizeCurrency(currency)

		coin, err := coinGecko.GetCoinDetail(coinID)
		if err != nil || coin.ID == "" {
			fmt.Printf("Error: Could not find coin with ID '%s'\n", coinID)
			os.Exit(1)
		}

		alert := models.Alert{
			CoinID:    coinID,
			Price:     price,
			Condition: condition,
			Currency:  currency,
		}

		if err := alertManager.AddAlert(alert); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s Alert added successfully!\n", titleColor("🔔"))
		fmt.Printf("You will be notified when %s goes %s %s%.2f\n",
			strings.ToUpper(coinID),
			condition,
			utils.CurrencySymbol(currency),
			price)
		fmt.Println("\nRun `crypto alert watch` (foreground) or `crypto alert start` (background) to monitor alerts.")
	},
}

var alertListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active alerts",
	Long: `Display all active price alerts.

OUTPUT INCLUDES:
  • Cryptocurrency name
  • Alert condition (above/below)
  • Target price
  • Creation date and time

EXAMPLE:
  crypto alert list`,
	Run: func(cmd *cobra.Command, args []string) {
		alerts := alertManager.GetAlerts()
		if len(alerts) == 0 {
			titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
			fmt.Printf("\n%s No active alerts\n", titleColor("🔔"))
			return
		}

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s Active Price Alerts\n\n", titleColor("🔔"))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Coin", "Condition", "Target Price", "Currency", "Created At"})
		table.SetBorder(false)
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
		)

		for _, alert := range alerts {
			currency := alert.Currency
			if currency == "" {
				currency = service.DEFAULT_CURRENCY
			}

			table.Rich([]string{
				strings.ToUpper(alert.CoinID),
				alert.Condition,
				fmt.Sprintf("%s%.2f", utils.CurrencySymbol(currency), alert.Price),
				strings.ToUpper(currency),
				alert.CreatedAt.Format("2006-01-02 15:04"),
			}, []tablewriter.Colors{
				{tablewriter.FgHiWhiteColor},
				{tablewriter.FgHiWhiteColor},
				{tablewriter.FgHiWhiteColor},
				{tablewriter.FgHiWhiteColor},
				{tablewriter.FgHiWhiteColor},
			})
		}

		table.Render()
	},
}

var alertRemoveCmd = &cobra.Command{
	Use:   "remove [coin-id] [price] [above/below]",
	Short: "Remove price alert",
	Long: `Remove price alerts for a cryptocurrency.

ARGUMENTS:
  coin-id  ID of the cryptocurrency (e.g., bitcoin)
  price    Optional target price (requires condition)
  type     Optional alert type: 'above' or 'below'

EXAMPLES:
  crypto alert remove bitcoin                    # Remove all alerts for Bitcoin
  crypto alert remove bitcoin 50000 above        # Remove specific alert`,
	Args: cobra.RangeArgs(1, 3),
	Run: func(cmd *cobra.Command, args []string) {
		coinID := utils.NormalizeCoinID(args[0])

		if len(args) == 1 {
			if err := alertManager.RemoveAlertsForCoin(coinID); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else if len(args) == 3 {
			price, err := strconv.ParseFloat(args[1], 64)
			if err != nil || price <= 0 {
				fmt.Println("Error: Invalid price value (must be greater than zero)")
				os.Exit(1)
			}
			condition := strings.ToLower(args[2])
			if condition != "above" && condition != "below" {
				fmt.Println("Error: Condition must be 'above' or 'below'")
				os.Exit(1)
			}
			if err := alertManager.RemoveAlertByTarget(coinID, price, condition); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Error: Provide coin-id only, or coin-id + price + condition")
			os.Exit(1)
		}

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s Alert removed successfully!\n", titleColor("🔔"))
	},
}

func init() {
	alertCmd.AddCommand(alertAddCmd)
	alertCmd.AddCommand(alertListCmd)
	alertCmd.AddCommand(alertRemoveCmd)
}
