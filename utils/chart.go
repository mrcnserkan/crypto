package utils

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/mrcnserkan/crypto/v2/models"
)

// ChartConfig controls terminal chart rendering.
type ChartConfig struct {
	Width          int
	Height         int
	CurrencySymbol string
	YTickCount     int
	XTickCount     int
}

// SeriesPoint is a single time-series data point for line charts.
type SeriesPoint struct {
	Time  time.Time
	Value float64
}

// DefaultChartConfig returns sensible defaults for terminal charts.
func DefaultChartConfig() ChartConfig {
	return ChartConfig{
		Width:          80,
		Height:         20,
		CurrencySymbol: "$",
		YTickCount:     5,
		XTickCount:     5,
	}
}

// ParseChartDate parses YYYY-MM-DD or YYYY-MM-DD HH:MM in UTC.
func ParseChartDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	layouts := []string{"2006-01-02 15:04", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format %q (use YYYY-MM-DD or YYYY-MM-DD HH:MM, UTC)", s)
}

// FilterSeriesByDateRange keeps points within [from, to] inclusive.
func FilterSeriesByDateRange(points []SeriesPoint, from, to *time.Time) []SeriesPoint {
	if from == nil && to == nil {
		return points
	}
	filtered := make([]SeriesPoint, 0, len(points))
	for _, p := range points {
		if from != nil && p.Time.Before(*from) {
			continue
		}
		if to != nil && p.Time.After(*to) {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}

// FilterOHLCByDateRange keeps candles within [from, to] inclusive.
func FilterOHLCByDateRange(data []models.OHLC, from, to *time.Time) []models.OHLC {
	if from == nil && to == nil {
		return data
	}
	filtered := make([]models.OHLC, 0, len(data))
	for _, c := range data {
		t := time.Unix(c.Time/1000, 0)
		if from != nil && t.Before(*from) {
			continue
		}
		if to != nil && t.After(*to) {
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered
}

// PeriodStats holds OHLC summary for a time range.
type PeriodStats struct {
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Change    float64
	ChangePct float64
}

// ComputePeriodStats calculates OHLC and change from series points.
func ComputePeriodStats(points []SeriesPoint) PeriodStats {
	if len(points) == 0 {
		return PeriodStats{}
	}
	open := points[0].Value
	close := points[len(points)-1].Value
	high, low := open, open
	for _, p := range points {
		if p.Value > high {
			high = p.Value
		}
		if p.Value < low {
			low = p.Value
		}
	}
	change := close - open
	changePct := 0.0
	if open != 0 {
		changePct = (change / open) * 100
	}
	return PeriodStats{
		Open: open, High: high, Low: low, Close: close,
		Change: change, ChangePct: changePct,
	}
}

// FormatPeriodStatsLine returns a compact OHLC summary string.
func FormatPeriodStatsLine(stats PeriodStats, symbol string) string {
	sign := "+"
	if stats.Change < 0 {
		sign = ""
	}
	return fmt.Sprintf(
		"O: %s%s  H: %s%s  L: %s%s  C: %s%s  Δ: %s%s%.2f (%s%.2f%%)",
		symbol, FormatCurrency(stats.Open),
		symbol, FormatCurrency(stats.High),
		symbol, FormatCurrency(stats.Low),
		symbol, FormatCurrency(stats.Close),
		sign, symbol, stats.Change,
		sign, stats.ChangePct,
	)
}

// RenderLineChart draws an ASCII line chart with aligned Y and X axes.
func RenderLineChart(points []SeriesPoint, cfg ChartConfig) string {
	if len(points) == 0 {
		return "No data available"
	}
	cfg = normalizeChartConfig(cfg)

	points = downsampleLTTB(points, cfg.Width)
	minP, maxP := seriesMinMax(points)
	minP, maxP = padPriceRange(minP, maxP)

	yLabels := buildYTickLabels(minP, maxP, cfg.YTickCount, cfg.CurrencySymbol)
	yAxisWidth := maxStringLen(yLabels)

	plot := newPlotGrid(cfg.Width, cfg.Height)
	for i := 1; i < len(points); i++ {
		x0 := indexToX(i-1, len(points), cfg.Width)
		y0 := priceToRow(points[i-1].Value, minP, maxP, cfg.Height)
		x1 := indexToX(i, len(points), cfg.Width)
		y1 := priceToRow(points[i].Value, minP, maxP, cfg.Height)
		drawLine(plot, x0, y0, x1, y1, '╱', '╲', '─')
	}
	for i, p := range points {
		x := indexToX(i, len(points), cfg.Width)
		y := priceToRow(p.Value, minP, maxP, cfg.Height)
		plot[y][x] = '●'
	}

	return assembleChart(plot, yLabels, yAxisWidth, cfg, points)
}

// RenderCandleChart draws an ASCII candlestick chart with Y and X axes.
func RenderCandleChart(data []models.OHLC, cfg ChartConfig) string {
	if len(data) == 0 {
		return "No data available"
	}
	cfg = normalizeChartConfig(cfg)

	candleWidth := 3
	maxCandles := cfg.Width / (candleWidth + 1)
	if len(data) > maxCandles {
		data = data[len(data)-maxCandles:]
	}

	var minP, maxP float64
	minP = data[0].Low
	maxP = data[0].High
	for _, c := range data {
		if c.Low < minP {
			minP = c.Low
		}
		if c.High > maxP {
			maxP = c.High
		}
	}
	minP, maxP = padPriceRange(minP, maxP)

	yLabels := buildYTickLabels(minP, maxP, cfg.YTickCount, cfg.CurrencySymbol)
	yAxisWidth := maxStringLen(yLabels)

	plot := newPlotGrid(cfg.Width, cfg.Height)
	for i, candle := range data {
		xCenter := i*(candleWidth+1) + candleWidth/2
		if xCenter >= cfg.Width {
			break
		}

		openY := priceToRow(candle.Open, minP, maxP, cfg.Height)
		closeY := priceToRow(candle.Close, minP, maxP, cfg.Height)
		highY := priceToRow(candle.High, minP, maxP, cfg.Height)
		lowY := priceToRow(candle.Low, minP, maxP, cfg.Height)

		bodyTop := intMin(openY, closeY)
		bodyBottom := intMax(openY, closeY)
		isBull := candle.Close >= candle.Open
		bodyChar := '░'
		if isBull {
			bodyChar = '█'
		}

		for y := highY; y < bodyTop; y++ {
			plot[y][xCenter] = '│'
		}
		for y := bodyBottom + 1; y <= lowY; y++ {
			plot[y][xCenter] = '│'
		}
		for y := bodyTop; y <= bodyBottom; y++ {
			for dx := 0; dx < candleWidth; dx++ {
				x := i*(candleWidth+1) + dx
				if x < cfg.Width {
					plot[y][x] = bodyChar
				}
			}
		}
	}

	series := ohlcToSeries(data)
	return assembleChart(plot, yLabels, yAxisWidth, cfg, series)
}

func normalizeChartConfig(cfg ChartConfig) ChartConfig {
	if cfg.Width < 20 {
		cfg.Width = 20
	}
	if cfg.Height < 8 {
		cfg.Height = 8
	}
	if cfg.YTickCount < 2 {
		cfg.YTickCount = 5
	}
	if cfg.XTickCount < 2 {
		cfg.XTickCount = 5
	}
	if cfg.CurrencySymbol == "" {
		cfg.CurrencySymbol = "$"
	}
	return cfg
}

func padPriceRange(minP, maxP float64) (float64, float64) {
	if minP == maxP {
		padding := math.Max(minP*0.01, 1)
		return minP - padding, maxP + padding
	}
	padding := (maxP - minP) * 0.05
	return minP - padding, maxP + padding
}

func buildYTickLabels(minP, maxP float64, tickCount int, symbol string) []string {
	labels := make([]string, tickCount)
	for i := 0; i < tickCount; i++ {
		ratio := float64(i) / float64(tickCount-1)
		price := maxP - ratio*(maxP-minP)
		labels[i] = symbol + FormatCurrency(price)
	}
	return labels
}

func maxStringLen(strs []string) int {
	maxLen := 0
	for _, s := range strs {
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	return maxLen
}

func newPlotGrid(width, height int) [][]rune {
	grid := make([][]rune, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]rune, width)
		for x := 0; x < width; x++ {
			grid[y][x] = ' '
		}
	}
	return grid
}

func priceToRow(price, minP, maxP float64, height int) int {
	if maxP == minP {
		return height / 2
	}
	ratio := (maxP - price) / (maxP - minP)
	y := int(math.Round(ratio * float64(height-1)))
	return clampInt(y, 0, height-1)
}

func indexToX(index, total, width int) int {
	if total <= 1 {
		return width / 2
	}
	return int(float64(index) / float64(total-1) * float64(width-1))
}

func drawLine(grid [][]rune, x0, y0, x1, y1 int, up, down, flat rune) {
	dx := intAbs(x1 - x0)
	dy := intAbs(y1 - y0)
	sx, sy := 1, 1
	if x0 >= x1 {
		sx = -1
	}
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy
	x, y := x0, y0
	for {
		if y >= 0 && y < len(grid) && x >= 0 && x < len(grid[0]) {
			ch := flat
			if y1 < y0 {
				ch = up
			} else if y1 > y0 {
				ch = down
			}
			if grid[y][x] == ' ' {
				grid[y][x] = ch
			}
		}
		if x == x1 && y == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func assembleChart(plot [][]rune, yLabels []string, yAxisWidth int, cfg ChartConfig, points []SeriesPoint) string {
	height := len(plot)
	width := len(plot[0])
	tickCount := len(yLabels)

	tickRows := make(map[int]int)
	for i := 0; i < tickCount; i++ {
		row := 0
		if tickCount > 1 {
			row = int(math.Round(float64(i) / float64(tickCount-1) * float64(height-1)))
		}
		tickRows[row] = i
	}

	var sb strings.Builder
	for row := 0; row < height; row++ {
		label := strings.Repeat(" ", yAxisWidth)
		if labelIdx, ok := tickRows[row]; ok {
			label = fmt.Sprintf("%*s", yAxisWidth, yLabels[labelIdx])
		}
		sb.WriteString(label)
		sb.WriteString(" ┤")
		for x := 0; x < width; x++ {
			sb.WriteRune(plot[row][x])
		}
		sb.WriteByte('\n')
	}

	sb.WriteString(strings.Repeat(" ", yAxisWidth))
	sb.WriteString(" └")
	sb.WriteString(strings.Repeat("─", width))
	sb.WriteByte('\n')

	xLabels := buildXTickLabels(points, cfg.XTickCount)
	if len(xLabels) > 0 {
		xLine := make([]rune, width)
		for i := range xLine {
			xLine[i] = ' '
		}
		for tickIdx, tick := range xLabels {
			label := formatXLabel(tick.time, points)
			pos := indexToX(tick.index, len(points), width)
			align := 0 // center
			if tickIdx == 0 {
				align = -1 // left
			} else if tickIdx == len(xLabels)-1 {
				align = 1 // right
			}
			placeXLabel(xLine, label, pos, width, align)
		}
		sb.WriteString(strings.Repeat(" ", yAxisWidth+2))
		sb.WriteString(string(xLine))
		sb.WriteByte('\n')
	}

	return sb.String()
}

type xTick struct {
	index int
	time  time.Time
}

func buildXTickLabels(points []SeriesPoint, tickCount int) []xTick {
	if len(points) == 0 {
		return nil
	}
	if len(points) == 1 {
		return []xTick{{index: 0, time: points[0].Time}}
	}
	ticks := make([]xTick, 0, tickCount)
	for i := 0; i < tickCount; i++ {
		idx := int(float64(i) / float64(tickCount-1) * float64(len(points)-1))
		ticks = append(ticks, xTick{index: idx, time: points[idx].Time})
	}
	return ticks
}

func formatXLabel(t time.Time, points []SeriesPoint) string {
	if len(points) < 2 {
		return t.Format("01-02")
	}
	span := points[len(points)-1].Time.Sub(points[0].Time)
	if span > 365*24*time.Hour {
		return t.Format("2006-01")
	}
	if span > 48*time.Hour {
		return t.Format("01-02")
	}
	return t.Format("15:04")
}

func placeXLabel(line []rune, label string, pos, width, align int) {
	start := pos - len(label)/2
	if align < 0 {
		start = pos
	} else if align > 0 {
		start = pos - len(label)
	}
	if start < 0 {
		start = 0
	}
	if start+len(label) > width {
		start = width - len(label)
	}
	if start < 0 {
		return
	}
	for i, ch := range label {
		x := start + i
		if x >= 0 && x < width {
			line[x] = ch
		}
	}
}

func downsampleLTTB(points []SeriesPoint, threshold int) []SeriesPoint {
	if len(points) <= threshold || threshold < 3 {
		return points
	}

	result := make([]SeriesPoint, 0, threshold)
	result = append(result, points[0])

	bucketSize := float64(len(points)-2) / float64(threshold-2)
	a := 0

	for i := 0; i < threshold-2; i++ {
		start := int(float64(i)*bucketSize) + 1
		end := int(float64(i+1)*bucketSize) + 1
		if end >= len(points) {
			end = len(points) - 1
		}

		rangeStart := int(float64(i+1)*bucketSize) + 1
		rangeEnd := int(float64(i+2)*bucketSize) + 1
		if rangeEnd >= len(points) {
			rangeEnd = len(points)
		}

		avgX, avgY := 0.0, 0.0
		count := rangeEnd - rangeStart
		if count <= 0 {
			continue
		}
		for j := rangeStart; j < rangeEnd; j++ {
			avgX += float64(j)
			avgY += points[j].Value
		}
		avgX /= float64(count)
		avgY /= float64(count)

		maxArea := -1.0
		nextIdx := start
		for j := start; j < end; j++ {
			area := math.Abs(
				(float64(a)-avgX)*(points[j].Value-points[a].Value) -
					(float64(a)-float64(j))*(avgY-points[a].Value),
			)
			if area > maxArea {
				maxArea = area
				nextIdx = j
			}
		}
		result = append(result, points[nextIdx])
		a = nextIdx
	}

	result = append(result, points[len(points)-1])
	return result
}

func downsampleSeries(points []SeriesPoint, maxPoints int) []SeriesPoint {
	if len(points) <= maxPoints || maxPoints < 2 {
		return points
	}
	result := make([]SeriesPoint, maxPoints)
	step := float64(len(points)-1) / float64(maxPoints-1)
	for i := 0; i < maxPoints; i++ {
		idx := int(math.Round(float64(i) * step))
		if idx >= len(points) {
			idx = len(points) - 1
		}
		result[i] = points[idx]
	}
	return result
}

func seriesMinMax(points []SeriesPoint) (float64, float64) {
	minP, maxP := points[0].Value, points[0].Value
	for _, p := range points[1:] {
		if p.Value < minP {
			minP = p.Value
		}
		if p.Value > maxP {
			maxP = p.Value
		}
	}
	return minP, maxP
}

func ohlcToSeries(data []models.OHLC) []SeriesPoint {
	series := make([]SeriesPoint, len(data))
	for i, c := range data {
		series[i] = SeriesPoint{
			Time:  time.Unix(c.Time/1000, 0),
			Value: c.Close,
		}
	}
	return series
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func intAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
