package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/mrcnserkan/crypto/models"
)

func TestParseChartDate(t *testing.T) {
	t1, err := ParseChartDate("2026-06-01")
	if err != nil || t1.Day() != 1 {
		t.Fatalf("ParseChartDate(date) = (%v, %v)", t1, err)
	}

	t2, err := ParseChartDate("2026-06-01 15:30")
	if err != nil || t2.Hour() != 15 || t2.Minute() != 30 {
		t.Fatalf("ParseChartDate(datetime) = (%v, %v)", t2, err)
	}

	_, err = ParseChartDate("bad")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestFilterSeriesByDateRange(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	points := []SeriesPoint{
		{Time: base, Value: 1},
		{Time: base.Add(24 * time.Hour), Value: 2},
		{Time: base.Add(48 * time.Hour), Value: 3},
	}
	from := base.Add(12 * time.Hour)
	to := base.Add(36 * time.Hour)

	filtered := FilterSeriesByDateRange(points, &from, &to)
	if len(filtered) != 1 || filtered[0].Value != 2 {
		t.Fatalf("FilterSeriesByDateRange() = %+v", filtered)
	}
}

func TestComputePeriodStats(t *testing.T) {
	stats := ComputePeriodStats([]SeriesPoint{
		{Value: 100},
		{Value: 120},
		{Value: 90},
		{Value: 110},
	})
	if stats.Open != 100 || stats.Close != 110 || stats.High != 120 || stats.Low != 90 {
		t.Fatalf("ComputePeriodStats() = %+v", stats)
	}
	if stats.ChangePct != 10 {
		t.Fatalf("ChangePct = %v, want 10", stats.ChangePct)
	}
}

func TestRenderLineChartHasAxes(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	points := make([]SeriesPoint, 10)
	for i := range points {
		points[i] = SeriesPoint{
			Time:  base.Add(time.Duration(i) * 24 * time.Hour),
			Value: 100 + float64(i)*5,
		}
	}
	out := RenderLineChart(points, ChartConfig{Width: 40, Height: 10, CurrencySymbol: "$", YTickCount: 3, XTickCount: 3})
	if !strings.Contains(out, "┤") {
		t.Fatal("line chart missing Y axis")
	}
	if !strings.Contains(out, "└") {
		t.Fatal("line chart missing X axis")
	}
	if strings.Count(out, "\n") < 12 {
		t.Fatal("line chart too short")
	}
}

func TestRenderCandleChartHasAxes(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC).Unix() * 1000
	data := []models.OHLC{
		{Time: base, Open: 100, High: 110, Low: 95, Close: 105},
		{Time: base + 86400000, Open: 105, High: 115, Low: 100, Close: 102},
		{Time: base + 172800000, Open: 102, High: 108, Low: 98, Close: 107},
	}
	out := RenderCandleChart(data, ChartConfig{Width: 30, Height: 10, CurrencySymbol: "$", YTickCount: 3, XTickCount: 3})
	if !strings.Contains(out, "┤") {
		t.Fatal("candle chart missing Y axis")
	}
	if !strings.Contains(out, "└") {
		t.Fatal("candle chart missing X axis")
	}
}

func TestPriceToRowMaxAtTop(t *testing.T) {
	if priceToRow(100, 0, 100, 10) != 0 {
		t.Fatal("max price should be top row")
	}
	if priceToRow(0, 0, 100, 10) != 9 {
		t.Fatal("min price should be bottom row")
	}
}

func TestFilterOHLCByDateRange(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC).Unix() * 1000
	data := []models.OHLC{
		{Time: base, Open: 100, High: 110, Low: 95, Close: 105},
		{Time: base + 86400000, Open: 105, High: 115, Low: 100, Close: 102},
		{Time: base + 172800000, Open: 102, High: 108, Low: 98, Close: 107},
	}
	from := time.UnixMilli(base + 43200000)
	to := time.UnixMilli(base + 129600000)

	filtered := FilterOHLCByDateRange(data, &from, &to)
	if len(filtered) != 1 || filtered[0].Close != 102 {
		t.Fatalf("FilterOHLCByDateRange() = %+v", filtered)
	}
}

func TestDownsampleLTTBPreservesEndpoints(t *testing.T) {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	points := make([]SeriesPoint, 100)
	for i := range points {
		points[i] = SeriesPoint{Time: base.Add(time.Duration(i) * time.Hour), Value: float64(i)}
	}
	out := downsampleLTTB(points, 10)
	if len(out) != 10 {
		t.Fatalf("expected 10 points, got %d", len(out))
	}
	if out[0].Value != 0 || out[len(out)-1].Value != 99 {
		t.Fatalf("LTTB should preserve first and last values, got %.0f and %.0f", out[0].Value, out[len(out)-1].Value)
	}
}
