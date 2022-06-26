/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	tableWriter "github.com/olekukonko/tablewriter"
)

func GetCellColorFromPriceChange(change float64) tableWriter.Colors {
	if change < 0 {
		return tableWriter.Colors{tableWriter.Normal, tableWriter.FgRedColor}
	}
	return tableWriter.Colors{tableWriter.Normal, tableWriter.FgGreenColor}
}

func ClearCoinName(coinName string) string {
	blankRegex := regexp.MustCompile(`[. ]`)
	specialCharRegex := regexp.MustCompile(`[(\[\])]`)
	return strings.ToLower(
		specialCharRegex.ReplaceAllString(
			blankRegex.ReplaceAllString(
				coinName, "-",
			), "",
		),
	)
}

func GetCaption() string {
	return fmt.Sprintf("Data source from coingecko.com at %s", currentTime())
}

func currentTime() string {
	currentTime := time.Now()
	return currentTime.Format("2006-01-02 15:04:05")
}
