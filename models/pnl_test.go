package models

import "testing"

func TestComputePortfolioPnL_WeightedAverage(t *testing.T) {
	p := NewPortfolio("")
	_ = p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 100, Currency: "usd", Type: "buy"})
	_ = p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 200, Currency: "usd", Type: "buy"})

	prices := map[string]float64{"bitcoin": 150}
	pnl := ComputePortfolioPnL(p, prices, "usd")

	if len(pnl.Coins) != 1 {
		t.Fatalf("expected 1 coin, got %d", len(pnl.Coins))
	}
	coin := pnl.Coins[0]
	if coin.AvgCost != 150 {
		t.Fatalf("AvgCost = %.2f, want 150", coin.AvgCost)
	}
	if coin.UnrealizedPnL != 0 {
		t.Fatalf("UnrealizedPnL = %.2f, want 0", coin.UnrealizedPnL)
	}
	if pnl.TotalValue != 300 {
		t.Fatalf("TotalValue = %.2f, want 300", pnl.TotalValue)
	}
}

func TestComputePortfolioPnL_RealizedOnSell(t *testing.T) {
	p := NewPortfolio("")
	_ = p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 2, Price: 100, Currency: "usd", Type: "buy"})
	_ = p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 150, Currency: "usd", Type: "sell"})

	prices := map[string]float64{"bitcoin": 120}
	pnl := ComputePortfolioPnL(p, prices, "usd")

	if len(pnl.Coins) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(pnl.Coins))
	}
	if pnl.TotalRealizedPnL != 50 {
		t.Fatalf("TotalRealizedPnL = %.2f, want 50", pnl.TotalRealizedPnL)
	}
}

func TestComputePortfolioPnL_MixedCurrency(t *testing.T) {
	p := NewPortfolio("")
	_ = p.AddTransaction(Transaction{CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 100, Currency: "eur", Type: "buy"})

	prices := map[string]float64{"bitcoin": 110}
	pnl := ComputePortfolioPnL(p, prices, "usd")

	if !pnl.HasMixedCurrency {
		t.Fatal("expected HasMixedCurrency true")
	}
}
