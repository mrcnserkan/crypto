/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

package utils

import (
	"regexp"
	"strings"

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
