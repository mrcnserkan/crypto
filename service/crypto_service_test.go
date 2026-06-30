package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mrcnserkan/crypto/v2/models"
)

func TestCoinGecko_SearchCoinsEncodesQuery(t *testing.T) {
	var receivedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query().Get("query")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"coins":[],"categories":[],"nfts":[]}`))
	}))
	defer server.Close()

	originalBaseURL := BaseURL
	t.Cleanup(func() { BaseURL = originalBaseURL })
	BaseURL = server.URL

	cg := NewCoinGecko()
	if _, err := cg.SearchCoins("bitcoin cash"); err != nil {
		t.Fatalf("SearchCoins() error = %v", err)
	}
	if receivedQuery != "bitcoin cash" {
		t.Fatalf("query = %q, want %q", receivedQuery, "bitcoin cash")
	}
}

func TestCoinGecko_GetRetriesOnRateLimit(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	originalBaseURL := BaseURL
	t.Cleanup(func() { BaseURL = originalBaseURL })
	BaseURL = server.URL

	cg := NewCoinGecko()
	coins, err := cg.GetMarkets("usd", 10, 1)
	if err != nil {
		t.Fatalf("GetMarkets() error = %v", err)
	}
	if coins == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if attempts < 2 {
		t.Fatalf("expected retry on 429, attempts = %d", attempts)
	}
}

func TestCoinGecko_GetMarketsByIDsBuildsRequest(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.String()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"bitcoin","symbol":"btc","name":"Bitcoin","current_price":100,"price_change_percentage_24h":1.2}]`))
	}))
	defer server.Close()

	originalBaseURL := BaseURL
	t.Cleanup(func() { BaseURL = originalBaseURL })
	BaseURL = server.URL

	cg := NewCoinGecko()
	coins, err := cg.GetMarketsByIDs("usd", []string{"bitcoin", "ethereum"})
	if err != nil {
		t.Fatalf("GetMarketsByIDs() error = %v", err)
	}
	if len(coins) != 1 {
		t.Fatalf("expected 1 coin, got %d", len(coins))
	}
	if !strings.Contains(receivedPath, "ids=bitcoin,ethereum") {
		t.Fatalf("unexpected request path: %s", receivedPath)
	}
}

func TestCoinGecko_GetMarketsByIDsChunksLargeRequests(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	originalBaseURL := BaseURL
	t.Cleanup(func() { BaseURL = originalBaseURL })
	BaseURL = server.URL

	ids := make([]string, 0, 301)
	for i := 0; i < 301; i++ {
		ids = append(ids, fmt.Sprintf("coin-%d", i))
	}

	cg := NewCoinGecko()
	if _, err := cg.GetMarketsByIDs("usd", ids); err != nil {
		t.Fatalf("GetMarketsByIDs() error = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("expected 2 chunked requests, got %d", requestCount)
	}
}

func TestAlertChecker_StopIsIdempotent(t *testing.T) {
	manager := models.NewAlertManager(t.TempDir())
	checker := NewAlertChecker(manager)

	checker.Start()
	checker.Stop()
	checker.Stop()
}

func TestCoinGecko_GetSimplePrices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/simple/price") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"bitcoin":{"usd":50000},"ethereum":{"usd":3000}}`))
	}))
	defer server.Close()

	originalBaseURL := BaseURL
	t.Cleanup(func() { BaseURL = originalBaseURL })
	BaseURL = server.URL

	cg := NewCoinGecko()
	prices, err := cg.GetSimplePrices([]string{"bitcoin", "ethereum"}, "usd")
	if err != nil {
		t.Fatalf("GetSimplePrices() error = %v", err)
	}
	if prices["bitcoin"] != 50000 || prices["ethereum"] != 3000 {
		t.Fatalf("unexpected prices: %+v", prices)
	}
}

func TestCoinGecko_GetRetriesOn5xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	originalBaseURL := BaseURL
	t.Cleanup(func() { BaseURL = originalBaseURL })
	BaseURL = server.URL

	cg := NewCoinGecko()
	if _, err := cg.GetMarkets("usd", 10, 1); err != nil {
		t.Fatalf("GetMarkets() error = %v", err)
	}
	if attempts < 2 {
		t.Fatalf("expected retry on 5xx, attempts = %d", attempts)
	}
}

func TestSelectIntervalForRange(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(10 * 24 * time.Hour)
	interval := SelectIntervalForRange(from, to)
	if interval.Days != 14 {
		t.Fatalf("expected 14d interval, got %+v", interval)
	}
}

func TestAlertChecker_TriggersAndRemovesAlert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"bitcoin":{"usd":51000}}`))
	}))
	defer server.Close()

	originalBaseURL := BaseURL
	t.Cleanup(func() { BaseURL = originalBaseURL })
	BaseURL = server.URL

	dir := t.TempDir()
	manager := models.NewAlertManager(dir)
	checker := NewAlertChecker(manager)
	checker.coinGecko = NewCoinGecko()

	_ = manager.AddAlert(models.Alert{CoinID: "bitcoin", Price: 50000, Condition: "above", Currency: "usd"})
	checker.RunOnce()

	if len(manager.GetAlerts()) != 0 {
		t.Fatalf("expected triggered alert removed, got %d alerts", len(manager.GetAlerts()))
	}
}
