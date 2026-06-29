package utils

import "testing"

func TestNormalizeCoinID(t *testing.T) {
	if got := NormalizeCoinID(" Bitcoin "); got != "bitcoin" {
		t.Fatalf("NormalizeCoinID() = %q, want bitcoin", got)
	}
}

func TestParsePage(t *testing.T) {
	page, err := ParsePage("2")
	if err != nil || page != 2 {
		t.Fatalf("ParsePage(2) = (%d, %v)", page, err)
	}

	_, err = ParsePage("0")
	if err == nil {
		t.Fatal("expected error for page 0")
	}

	_, err = ParsePage("abc")
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestParsePerPage(t *testing.T) {
	perPage, err := ParsePerPage("250")
	if err != nil || perPage != 250 {
		t.Fatalf("ParsePerPage(250) = (%d, %v)", perPage, err)
	}

	_, err = ParsePerPage("251")
	if err == nil {
		t.Fatal("expected error for per-page above max")
	}
}

func TestFormatISODate(t *testing.T) {
	if got := FormatISODate("2025-10-06T18:57:42.558Z"); got != "2025-10-06" {
		t.Fatalf("FormatISODate() = %q", got)
	}
	if got := FormatISODate("short"); got != "short" {
		t.Fatalf("FormatISODate(short) = %q", got)
	}
}

func TestCurrencySymbol(t *testing.T) {
	if got := CurrencySymbol("usd"); got != "$" {
		t.Fatalf("CurrencySymbol(usd) = %q", got)
	}
	if got := CurrencySymbol("eur"); got != "EUR" {
		t.Fatalf("CurrencySymbol(eur) = %q", got)
	}
}

func TestIsEffectivelyZero(t *testing.T) {
	if !IsEffectivelyZero(5.551115123125783e-17) {
		t.Fatal("expected float dust to be effectively zero")
	}
	if IsEffectivelyZero(0.0001) {
		t.Fatal("expected small but real holding to remain")
	}
}

func TestPriceFromCurrencyMap(t *testing.T) {
	prices := map[string]float64{"usd": 100, "eur": 90}

	price, err := PriceFromCurrencyMap(prices, "usd")
	if err != nil || price != 100 {
		t.Fatalf("PriceFromCurrencyMap(usd) = (%v, %v)", price, err)
	}

	_, err = PriceFromCurrencyMap(prices, "invalid")
	if err == nil {
		t.Fatal("expected error for invalid currency")
	}
}

func TestFormatCurrency(t *testing.T) {
	if got := FormatCurrency(1500); got != "1.50K" {
		t.Fatalf("FormatCurrency(1500) = %q", got)
	}
}
