package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPortfolio_FloatDustRemovedAfterFullSell(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "portfolio.json")
	p := NewPortfolio(filePath)

	for _, buyAmount := range []float64{0.1, 0.2} {
		if err := p.AddTransaction(Transaction{
			CoinID: "bitcoin", Symbol: "btc", Amount: buyAmount, Price: 100, Currency: "usd", Type: "buy",
		}); err != nil {
			t.Fatalf("AddTransaction(buy) error = %v", err)
		}
	}

	if err := p.AddTransaction(Transaction{
		CoinID: "bitcoin", Symbol: "btc", Amount: 0.3, Price: 100, Currency: "usd", Type: "sell",
	}); err != nil {
		t.Fatalf("AddTransaction(sell) error = %v", err)
	}

	if !p.HasHoldings() {
		return
	}
	t.Fatalf("expected dust holding to be removed, got %+v", p.Holdings)
}

func TestPortfolio_LoadPersistsNormalization(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "portfolio.json")
	payload := `{"holdings":{"BITCOIN":1},"transactions":[{"coin_id":"BITCOIN","symbol":"btc","amount":1,"price":100,"type":"buy","date":"2024-01-01T00:00:00Z"}]}`
	if err := os.WriteFile(filePath, []byte(payload), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	p := NewPortfolio(filePath)
	if err := p.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	stored, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(stored), `"bitcoin"`) {
		t.Fatalf("expected normalized coin id in saved file, got %s", string(stored))
	}
}

func TestPortfolio_AddTransactionBuyAndSell(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "portfolio.json")
	p := NewPortfolio(filePath)

	buy := Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 100, Currency: "usd", Type: "buy"}
	if err := p.AddTransaction(buy); err != nil {
		t.Fatalf("AddTransaction(buy) error = %v", err)
	}
	if p.GetHolding("bitcoin") != 1 {
		t.Fatalf("expected holding 1, got %.6f", p.GetHolding("bitcoin"))
	}

	sell := Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 120, Currency: "usd", Type: "sell"}
	if err := p.AddTransaction(sell); err != nil {
		t.Fatalf("AddTransaction(sell) error = %v", err)
	}
	if _, exists := p.Holdings["bitcoin"]; exists {
		t.Fatal("expected zero holding to be removed from map")
	}
}

func TestPortfolio_InsufficientBalance(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "portfolio.json")
	p := NewPortfolio(filePath)

	err := p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 100, Currency: "usd", Type: "sell"})
	if err == nil {
		t.Fatal("expected insufficient balance error")
	}
}

func TestPortfolio_Validation(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "portfolio.json")
	p := NewPortfolio(filePath)

	err := p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: -1, Price: 100, Currency: "usd", Type: "buy"})
	if err == nil {
		t.Fatal("expected error for negative amount")
	}

	err = p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 0, Currency: "usd", Type: "buy"})
	if err == nil {
		t.Fatal("expected error for zero price")
	}
}

func TestPortfolio_LoadNormalizesHoldings(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "portfolio.json")
	payload := `{"holdings":{"BITCOIN":1,"bitcoin":0.5},"transactions":[{"coin_id":"BITCOIN","symbol":"btc","amount":1.5,"price":100,"type":"buy","date":"2024-01-01T00:00:00Z"}]}`
	if err := os.WriteFile(filePath, []byte(payload), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	p := NewPortfolio(filePath)
	if err := p.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if p.GetHolding("bitcoin") != 1.5 {
		t.Fatalf("expected merged holding 1.5, got %.6f", p.GetHolding("bitcoin"))
	}
	if p.Transactions[0].CoinID != "bitcoin" {
		t.Fatalf("expected normalized transaction coin id, got %s", p.Transactions[0].CoinID)
	}
}

func TestPortfolio_SaveUsesPrivatePermissions(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "portfolio.json")
	p := NewPortfolio(filePath)

	if err := p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 100, Currency: "usd", Type: "buy"}); err != nil {
		t.Fatalf("AddTransaction() error = %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected file mode 0600, got %#o", info.Mode().Perm())
	}
}

func TestPortfolio_HasAnyDataAfterFullSell(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "portfolio.json")
	p := NewPortfolio(filePath)

	_ = p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 100, Currency: "usd", Type: "buy"})
	_ = p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 120, Currency: "usd", Type: "sell"})

	if p.HasHoldings() {
		t.Fatal("expected no holdings after full sell")
	}
	if !p.HasAnyData() {
		t.Fatal("expected HasAnyData true when transaction history remains")
	}

	if err := p.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	if p.HasAnyData() {
		t.Fatal("expected empty portfolio after Clear()")
	}
}

func TestPortfolio_RemoveCoinWithHistoryOnly(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "portfolio.json")
	p := NewPortfolio(filePath)

	_ = p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 100, Currency: "usd", Type: "buy"})
	_ = p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 120, Currency: "usd", Type: "sell"})

	if !p.HasCoinData("bitcoin") {
		t.Fatal("expected HasCoinData true for history-only coin")
	}
	if err := p.RemoveCoin("bitcoin"); err != nil {
		t.Fatalf("RemoveCoin() error = %v", err)
	}
	if p.HasCoinData("bitcoin") {
		t.Fatal("expected coin data removed")
	}
}
