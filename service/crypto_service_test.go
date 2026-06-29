package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrcnserkan/crypto/models"
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

func TestAlertChecker_StopIsIdempotent(t *testing.T) {
	manager := models.NewAlertManager(t.TempDir())
	checker := NewAlertChecker(manager)

	checker.Start()
	checker.Stop()
	checker.Stop()
}
