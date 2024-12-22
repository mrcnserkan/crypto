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
	"github.com/mrcnserkan/crypto/models"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var alertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Manage price alerts",
	Long: `Set and manage price alerts for cryptocurrencies.
Alerts will notify you when a coin reaches a specified price.

Note: Alerts are checked every 5 minutes to avoid API rate limits.
Alerts are automatically removed once triggered.`,
}

var alertAddCmd = &cobra.Command{
	Use:   "add [coin-id] [price] [above/below]",
	Short: "Add a new price alert",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		coinID := args[0]
		price, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			fmt.Println("Error: Invalid price value")
			return
		}
		condition := strings.ToLower(args[2])
		if condition != "above" && condition != "below" {
			fmt.Println("Error: Condition must be 'above' or 'below'")
			return
		}

		// Verify coin exists
		coin, err := coinGecko.GetCoinDetail(coinID)
		if err != nil || coin.ID == "" {
			fmt.Printf("Error: Could not find coin with ID '%s'\n", coinID)
			return
		}

		alert := models.Alert{
			CoinID:    coinID,
			Price:     price,
			Condition: condition,
		}

		if err := alertManager.AddAlert(alert); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s Alert added successfully!\n", titleColor("🔔"))
		fmt.Printf("You will be notified when %s goes %s $%.2f\n",
			strings.ToUpper(coinID),
			condition,
			price)
	},
}

var alertListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active alerts",
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
		table.SetHeader([]string{"Coin", "Condition", "Target Price", "Created At"})
		table.SetBorder(false)
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
		)

		for _, alert := range alerts {
			table.Rich([]string{
				strings.ToUpper(alert.CoinID),
				alert.Condition,
				fmt.Sprintf("$%.2f", alert.Price),
				alert.CreatedAt.Format("2006-01-02 15:04"),
			}, []tablewriter.Colors{
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
	Use:   "remove [coin-id]",
	Short: "Remove an alert",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		coinID := args[0]
		if err := alertManager.RemoveAlert(coinID); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
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
