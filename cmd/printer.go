/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrcnserkan/crypto/service"
	"github.com/mrcnserkan/crypto/utils"
	tableWriter "github.com/olekukonko/tablewriter"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func PrintList(page string, perPage string, currency string) {
	currencySymbol := service.DEFAULT_CURRENCY_SYMBOL
	if currency != service.DEFAULT_CURRENCY {
		currencySymbol = strings.ToUpper(currency)
	}
	printer := message.NewPrinter(language.English)
	table := tableWriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Coin", "", "Price", "24h", "7d", "Market Cap", "Ath"})
	table.SetBorder(false)
	table.SetRowLine(true)
	for _, element := range service.GetList(page, perPage, currency) {
		table.Rich([]string{
			fmt.Sprintf("%d", element.MarketCapRank),
			element.Name,
			strings.ToUpper((element.Symbol)),
			printer.Sprintf("%s%.2f", currencySymbol, element.CurrentPrice),
			fmt.Sprintf("%.1f%%", element.PriceChangePercentage24HInCurrency),
			fmt.Sprintf("%.1f%%", element.PriceChangePercentage7DInCurrency),
			printer.Sprintf("%s%d", currencySymbol, element.MarketCap),
			printer.Sprintf("%s%.2f", currencySymbol, element.Ath),
		}, []tableWriter.Colors{
			{},
			{},
			{},
			{tableWriter.Bold},
			utils.GetCellColorFromPriceChange(element.PriceChangePercentage24HInCurrency),
			utils.GetCellColorFromPriceChange(element.PriceChangePercentage7DInCurrency),
			{},
			{},
		})
	}
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
	coinDetail := service.GetDetail(coinName)
	printer := message.NewPrinter(language.English)
	table := tableWriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Coin", "", "Price", "24h", "7d", "Market Cap", "Ath"})
	table.SetBorder(false)
	table.SetRowLine(true)
	table.Rich([]string{
		fmt.Sprintf("%d", coinDetail.MarketCapRank),
		coinDetail.Name,
		strings.ToUpper((coinDetail.Symbol)),
		printer.Sprintf("%s%.2f", currencySymbol, coinDetail.MarketData.CurrentPrice[currency]),
		fmt.Sprintf("%.1f%%", coinDetail.MarketData.PriceChangePercentage24HInCurrency[currency]),
		fmt.Sprintf("%.1f%%", coinDetail.MarketData.PriceChangePercentage7DInCurrency[currency]),
		printer.Sprintf("%s%d", currencySymbol, coinDetail.MarketData.MarketCap[currency]),
		printer.Sprintf("%s%.2f", currencySymbol, coinDetail.MarketData.Ath[currency]),
	}, []tableWriter.Colors{
		{},
		{},
		{},
		{tableWriter.Bold},
		utils.GetCellColorFromPriceChange(coinDetail.MarketData.PriceChangePercentage24HInCurrency[currency]),
		utils.GetCellColorFromPriceChange(coinDetail.MarketData.PriceChangePercentage24HInCurrency[currency]),
		{},
		{},
	})
	table.Render()
}

func PrintSearchResult(query string) {
	table := tableWriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Market Cap Rank", "Id", "Name", "Symbol"})
	table.SetBorder(false)
	table.SetRowLine(true)
	for _, element := range service.Search(query).Coins {
		table.Append([]string{
			fmt.Sprintf("%d", element.MarketCapRank),
			element.ID,
			element.Name,
			element.Symbol,
		})
	}
	table.Render()
}
