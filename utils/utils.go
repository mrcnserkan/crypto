/*
Copyright © 2022 Serkan MERCAN <serkanmercan@email.com>

*/

package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mrcnserkan/crypto/models"
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
	return fmt.Sprintf("Data source from coingecko.com at %s", GetCurrentTime())
}

func GetCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func StringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 1 // Default value
	}
	return i
}

func FormatCurrency(value float64) string {
	if value >= 1_000_000_000_000 { // Trillion
		return fmt.Sprintf("%.2fT", value/1_000_000_000_000)
	} else if value >= 1_000_000_000 { // Billion
		return fmt.Sprintf("%.2fB", value/1_000_000_000)
	} else if value >= 1_000_000 { // Million
		return fmt.Sprintf("%.2fM", value/1_000_000)
	} else if value >= 1_000 { // Thousand
		return fmt.Sprintf("%.2fK", value/1_000)
	}
	return fmt.Sprintf("%.2f", value)
}

func MinMax(array []float64) (float64, float64) {
	if len(array) == 0 {
		return 0, 0
	}
	min := array[0]
	max := array[0]
	for _, value := range array {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	return min, max
}

func GenerateCandleChart(data []models.OHLC) string {
	if len(data) == 0 {
		return "No data available"
	}

	// Chart height and width
	height := 20
	width := 100

	// Find highest and lowest prices
	var maxPrice, minPrice float64
	maxPrice = data[0].High
	minPrice = data[0].Low
	for _, candle := range data {
		if candle.High > maxPrice {
			maxPrice = candle.High
		}
		if candle.Low < minPrice {
			minPrice = candle.Low
		}
	}

	// Price range
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1 // Prevent division by zero
	}

	// Width for each candle
	candleWidth := 5
	maxCandles := width / candleWidth
	if len(data) > maxCandles {
		data = data[len(data)-maxCandles:]
	}

	// Create chart matrix (extra space for y-axis)
	chart := make([][]string, height+1) // +1 for x-axis
	for i := range chart {
		chart[i] = make([]string, width+12) // +12 for y-axis labels
		for j := range chart[i] {
			chart[i][j] = " "
		}
	}

	// Add Y-axis labels
	priceStep := priceRange / float64(height-1)
	for i := 0; i < height; i++ {
		price := maxPrice - (float64(i) * priceStep)
		priceLabel := fmt.Sprintf("%8.0f │", price)
		for j, r := range priceLabel {
			chart[i][j] = string(r)
		}
	}

	// For each candle
	for i, candle := range data {
		x := i*candleWidth + 12 // offset for y-axis labels
		if x >= width+12 {
			break
		}

		// Calculate candle positions
		openY := int(((candle.Open - minPrice) / priceRange) * float64(height-1))
		closeY := int(((candle.Close - minPrice) / priceRange) * float64(height-1))
		highY := int(((candle.High - minPrice) / priceRange) * float64(height-1))
		lowY := int(((candle.Low - minPrice) / priceRange) * float64(height-1))

		// Determine if bullish or bearish
		isGreen := candle.Close >= candle.Open

		// Draw wicks (highest and lowest points)
		// Upper wick
		maxY := highY
		bodyTop := max(openY, closeY)
		for y := bodyTop + 1; y <= maxY; y++ {
			if y >= 0 && y < height {
				chart[y][x+2] = "│"
			}
		}

		// Lower wick
		minY := lowY
		bodyBottom := min(openY, closeY)
		for y := minY; y < bodyBottom; y++ {
			if y >= 0 && y < height {
				chart[y][x+2] = "│"
			}
		}

		// Draw candle body
		startY := min(openY, closeY)
		endY := max(openY, closeY)

		for y := startY; y <= endY; y++ {
			if y >= 0 && y < height {
				if isGreen {
					chart[y][x+1] = "█"
					chart[y][x+2] = "█"
					chart[y][x+3] = "█"
				} else {
					chart[y][x+1] = "░"
					chart[y][x+2] = "░"
					chart[y][x+3] = "░"
				}
			}
		}
	}

	// Add X-axis line
	for i := 12; i < width+12; i++ {
		chart[height-1][i] = "─"
	}

	// Convert chart to string
	var result strings.Builder
	for _, row := range chart {
		result.WriteString(strings.Join(row, "") + "\n")
	}

	return result.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
