package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mrcnserkan/crypto/v2/models"
	"github.com/spf13/cobra"
)

var portfolioExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export portfolio to CSV or JSON",
	Long: `Export current portfolio holdings with P&L data.

EXAMPLES:
  crypto portfolio export
  crypto portfolio export --format json --output portfolio.json`,
	Run: func(cmd *cobra.Command, args []string) {
		if !portfolio.HasHoldings() {
			fmt.Println("Portfolio is empty")
			return
		}

		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		currency := getCurrencyFlag(cmd)

		coinIDs := make([]string, 0, len(portfolio.Holdings))
		for coinID := range portfolio.Holdings {
			coinIDs = append(coinIDs, coinID)
		}

		coins, err := coinGecko.GetMarketsByIDs(currency, coinIDs)
		if err != nil {
			fmt.Printf("Error fetching market data: %v\n", err)
			os.Exit(1)
		}

		prices := make(map[string]float64, len(coins))
		names := make(map[string]string, len(coins))
		for _, coin := range coins {
			prices[coin.ID] = coin.CurrentPrice
			names[coin.ID] = coin.Name
		}

		pnl := models.ComputePortfolioPnL(portfolio, prices, currency)

		switch strings.ToLower(format) {
		case "json":
			if err := exportPortfolioJSON(pnl, names, output); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		default:
			if err := exportPortfolioCSV(pnl, names, currency, output); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

		if output != "" {
			fmt.Printf("Portfolio exported to %s\n", output)
		}
	},
}

func exportPortfolioCSV(pnl models.PortfolioPnL, names map[string]string, currency, output string) error {
	var writer *csv.Writer
	if output == "" {
		writer = csv.NewWriter(os.Stdout)
	} else {
		file, err := os.Create(output)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = csv.NewWriter(file)
	}

	_ = writer.Write([]string{"coin_id", "name", "amount", "avg_cost", "current_price", "value", "pnl", "pnl_pct", "currency"})
	for _, coin := range pnl.Coins {
		name := names[coin.CoinID]
		if name == "" {
			name = coin.CoinID
		}
		_ = writer.Write([]string{
			coin.CoinID,
			name,
			fmt.Sprintf("%.8f", coin.Amount),
			fmt.Sprintf("%.2f", coin.AvgCost),
			fmt.Sprintf("%.2f", coin.CurrentPrice),
			fmt.Sprintf("%.2f", coin.CurrentValue),
			fmt.Sprintf("%.2f", coin.UnrealizedPnL),
			fmt.Sprintf("%.2f", coin.UnrealizedPnLPct),
			currency,
		})
	}
	writer.Flush()
	return writer.Error()
}

func exportPortfolioJSON(pnl models.PortfolioPnL, names map[string]string, output string) error {
	type row struct {
		CoinID       string  `json:"coin_id"`
		Name         string  `json:"name"`
		Amount       float64 `json:"amount"`
		AvgCost      float64 `json:"avg_cost"`
		CurrentPrice float64 `json:"current_price"`
		Value        float64 `json:"value"`
		PnL          float64 `json:"pnl"`
		PnLPct       float64 `json:"pnl_pct"`
	}
	rows := make([]row, 0, len(pnl.Coins))
	for _, coin := range pnl.Coins {
		name := names[coin.CoinID]
		if name == "" {
			name = coin.CoinID
		}
		rows = append(rows, row{
			CoinID: coin.CoinID, Name: name, Amount: coin.Amount,
			AvgCost: coin.AvgCost, CurrentPrice: coin.CurrentPrice,
			Value: coin.CurrentValue, PnL: coin.UnrealizedPnL, PnLPct: coin.UnrealizedPnLPct,
		})
	}

	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	if output == "" {
		fmt.Println(string(data))
		return nil
	}
	return os.WriteFile(output, data, 0o600)
}

func init() {
	portfolioExportCmd.Flags().String("format", "csv", "Export format: csv or json")
	portfolioExportCmd.Flags().String("output", "", "Output file path (stdout if empty)")
	portfolioCmd.AddCommand(portfolioExportCmd)
}
