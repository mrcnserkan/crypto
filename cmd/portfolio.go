package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"os"

	"github.com/fatih/color"
	"github.com/mrcnserkan/crypto/models"
	"github.com/mrcnserkan/crypto/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var portfolioCmd = &cobra.Command{
	Use:   "portfolio",
	Short: "Portfolio management",
	Long: `Manage your cryptocurrency portfolio and track your transactions.
	
Features:
- Add buy/sell transactions with price history
- Track portfolio value in real-time
- View detailed transaction history
- Remove specific coins or clear entire portfolio
- Support for multiple currencies`,
}

var portfolioAddCmd = &cobra.Command{
	Use:   "add [coin-id] [amount] [price] [buy/sell]",
	Short: "Add transaction to portfolio",
	Long: `Add a buy or sell transaction to your portfolio.
	
Arguments:
  coin-id: ID of the cryptocurrency (e.g., bitcoin, ethereum)
  amount:  Amount of coins in the transaction
  price:   Price per coin at the time of transaction
  type:    Transaction type (buy or sell)

Example:
  crypto portfolio add bitcoin 0.5 50000 buy    # Buy 0.5 BTC at $50,000
  crypto portfolio add ethereum 2.0 3000 sell   # Sell 2.0 ETH at $3,000`,
	Args: cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		coinID := args[0]
		amount, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			fmt.Println("Error: Invalid amount")
			return
		}
		price, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			fmt.Println("Error: Invalid price")
			return
		}
		transactionType := args[3]
		if transactionType != "buy" && transactionType != "sell" {
			fmt.Println("Error: Transaction type must be 'buy' or 'sell'")
			return
		}

		// Get coin details from CoinGecko
		coin, err := coinGecko.GetCoinDetail(coinID)
		if err != nil {
			fmt.Printf("Error: Could not verify coin ID: %v\n", err)
			return
		}

		transaction := models.Transaction{
			CoinID: coinID,
			Symbol: coin.Symbol,
			Amount: amount,
			Price:  price,
			Type:   transactionType,
			Date:   time.Now(),
		}

		if err := portfolio.AddTransaction(transaction); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s %s transaction added successfully!\n", titleColor("💰"), titleColor(fmt.Sprintf("%s %.6f %s @ %.2f",
			strings.ToUpper(transactionType),
			amount,
			strings.ToUpper(coin.Symbol),
			price)))
	},
}

var portfolioListCmd = &cobra.Command{
	Use:   "list",
	Short: "View portfolio status",
	Long: `Display current portfolio holdings with real-time values and 24h changes.
	
The output includes:
- Coin name and symbol
- Amount held
- Current price
- Total value
- 24h price change percentage
- Total portfolio value

Flags:
  --currency string   Currency for portfolio valuation (default "usd")

Example:
  crypto portfolio list
  crypto portfolio list --currency eur`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(portfolio.Holdings) == 0 {
			titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
			fmt.Printf("\n%s %s\n", titleColor("💼"), titleColor("Portfolio is empty"))
			return
		}

		currency, _ := cmd.Flags().GetString("currency")
		currencySymbol := "$"
		if currency != "usd" {
			currencySymbol = strings.ToUpper(currency)
		}

		prices := make(map[string]float64)
		totalValue := 0.0

		// Color definitions
		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()

		fmt.Printf("\n%s %s\n\n", titleColor("💼"), titleColor("Portfolio Holdings"))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Coin", "Amount", "Price", "Value", "24h Change"})
		table.SetBorder(false)
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
		)

		for coinID, amount := range portfolio.Holdings {
			coin, err := coinGecko.GetCoinDetail(coinID)
			if err != nil {
				fmt.Printf("Error fetching price for %s: %v\n", coinID, err)
				continue
			}

			currentPrice := coin.MarketData.CurrentPrice[currency]
			value := amount * currentPrice
			prices[coinID] = currentPrice
			totalValue += value
			change24h := coin.MarketData.PriceChangePercentage24HInCurrency[currency]

			table.Rich([]string{
				fmt.Sprintf("%s (%s)", coin.Name, strings.ToUpper(coin.Symbol)),
				fmt.Sprintf("%.6f", amount),
				fmt.Sprintf("%s%s", currencySymbol, utils.FormatCurrency(currentPrice)),
				fmt.Sprintf("%s%s", currencySymbol, utils.FormatCurrency(value)),
				fmt.Sprintf("%.2f%%", change24h),
			}, []tablewriter.Colors{
				{tablewriter.FgHiWhiteColor},
				{tablewriter.FgHiWhiteColor},
				{tablewriter.FgHiWhiteColor},
				{tablewriter.FgHiWhiteColor},
				utils.GetCellColorFromPriceChange(change24h),
			})
		}

		table.SetFooter([]string{
			"Total Value",
			"",
			"",
			fmt.Sprintf("%s%s", currencySymbol, utils.FormatCurrency(totalValue)),
			"",
		})
		table.SetFooterColor(
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
			nil,
			nil,
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
			nil,
		)

		table.Render()
	},
}

var portfolioHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "View transaction history",
	Long: `Display a chronological list of all portfolio transactions.
	
The output includes:
- Transaction date and time
- Transaction type (BUY/SELL)
- Coin details
- Amount traded
- Price at transaction

Flags:
  --currency string   Currency for price display (default "usd")

Example:
  crypto portfolio history
  crypto portfolio history --currency eur`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(portfolio.Transactions) == 0 {
			titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
			fmt.Printf("\n%s %s\n", titleColor("📜"), titleColor("No transaction history"))
			return
		}

		currency, _ := cmd.Flags().GetString("currency")
		currencySymbol := "$"
		if currency != "usd" {
			currencySymbol = strings.ToUpper(currency)
		}

		// Color definitions
		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()

		fmt.Printf("\n%s %s\n\n", titleColor("📜"), titleColor("Transaction History"))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "Type", "Coin", "Amount", "Price"})
		table.SetBorder(false)
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgHiBlueColor},
		)

		for _, t := range portfolio.Transactions {
			typeColor := tablewriter.FgGreenColor
			if t.Type == "sell" {
				typeColor = tablewriter.FgRedColor
			}

			table.Rich([]string{
				t.Date.Format("2006-01-02 15:04"),
				strings.ToUpper(t.Type),
				fmt.Sprintf("%s (%s)", t.CoinID, strings.ToUpper(t.Symbol)),
				fmt.Sprintf("%.6f", t.Amount),
				fmt.Sprintf("%s%s", currencySymbol, utils.FormatCurrency(t.Price)),
			}, []tablewriter.Colors{
				{tablewriter.FgHiWhiteColor},
				{typeColor},
				{tablewriter.FgHiWhiteColor},
				{tablewriter.FgHiWhiteColor},
				{tablewriter.FgHiWhiteColor},
			})
		}

		table.Render()
	},
}

var portfolioClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear entire portfolio",
	Long: `Remove all coins and transactions from your portfolio.
	
This command will:
- Remove all cryptocurrency holdings
- Delete all transaction history
- Require confirmation before proceeding
- Cannot be undone

Example:
  crypto portfolio clear`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(portfolio.Holdings) == 0 {
			titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
			fmt.Printf("\n%s %s\n", titleColor("💼"), titleColor("Portfolio is already empty"))
			return
		}

		// Ask for confirmation
		fmt.Print("\nAre you sure you want to clear your entire portfolio? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Operation cancelled")
			return
		}

		portfolio.Holdings = make(map[string]float64)
		portfolio.Transactions = []models.Transaction{}
		if err := portfolio.Save(); err != nil {
			fmt.Printf("Error clearing portfolio: %v\n", err)
			return
		}

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s %s\n", titleColor("💼"), titleColor("Portfolio cleared successfully"))
	},
}

var portfolioRemoveCmd = &cobra.Command{
	Use:   "remove [coin-id]",
	Short: "Remove a coin from portfolio",
	Long: `Remove a specific coin and all its transactions from your portfolio.
	
This command will:
- Remove the specified coin from holdings
- Delete all transactions for this coin
- Require confirmation before proceeding
- Cannot be undone

Arguments:
  coin-id: ID of the cryptocurrency to remove (e.g., bitcoin)

Example:
  crypto portfolio remove bitcoin`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		coinID := args[0]

		// Check if coin exists in portfolio
		amount, exists := portfolio.Holdings[coinID]
		if !exists {
			fmt.Printf("Error: %s is not in your portfolio\n", coinID)
			return
		}

		// Ask for confirmation
		fmt.Printf("\nAre you sure you want to remove %s (Amount: %.6f) from your portfolio? (y/N): ",
			strings.ToUpper(coinID), amount)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Operation cancelled")
			return
		}

		// Remove coin from holdings
		delete(portfolio.Holdings, coinID)

		// Remove all transactions for this coin
		newTransactions := []models.Transaction{}
		for _, t := range portfolio.Transactions {
			if t.CoinID != coinID {
				newTransactions = append(newTransactions, t)
			}
		}
		portfolio.Transactions = newTransactions

		if err := portfolio.Save(); err != nil {
			fmt.Printf("Error removing coin: %v\n", err)
			return
		}

		titleColor := color.New(color.FgHiCyan, color.Bold).SprintFunc()
		fmt.Printf("\n%s %s removed from portfolio successfully\n",
			titleColor("💼"), titleColor(strings.ToUpper(coinID)))
	},
}

func init() {
	portfolioCmd.PersistentFlags().String("currency", "usd", "Currency for portfolio valuation")
	portfolioCmd.AddCommand(portfolioAddCmd)
	portfolioCmd.AddCommand(portfolioListCmd)
	portfolioCmd.AddCommand(portfolioHistoryCmd)
	portfolioCmd.AddCommand(portfolioClearCmd)
	portfolioCmd.AddCommand(portfolioRemoveCmd)
}
