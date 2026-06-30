package models

import "strings"

// CoinPnL holds weighted-average cost basis and P&L for a single coin.
type CoinPnL struct {
	CoinID            string
	Amount            float64
	AvgCost           float64
	TotalCost         float64
	CurrentPrice      float64
	CurrentValue      float64
	UnrealizedPnL     float64
	UnrealizedPnLPct  float64
	RealizedPnL       float64
	HasMixedCurrency  bool
	TransactionCurrency string
}

// PortfolioPnL aggregates P&L across all holdings.
type PortfolioPnL struct {
	Coins              []CoinPnL
	TotalValue         float64
	TotalCost          float64
	TotalUnrealizedPnL float64
	TotalRealizedPnL   float64
	HasMixedCurrency   bool
}

// ComputePortfolioPnL calculates weighted-average cost and unrealized P&L.
func ComputePortfolioPnL(p *Portfolio, prices map[string]float64, displayCurrency string) PortfolioPnL {
	displayCurrency = normalizeCurrency(displayCurrency)
	result := PortfolioPnL{}

	coinIDs := make(map[string]struct{})
	for coinID := range p.Holdings {
		coinIDs[coinID] = struct{}{}
	}

	type costState struct {
		amount      float64
		totalCost   float64
		realizedPnL float64
		currency    string
		mixed       bool
	}
	states := make(map[string]*costState)

	for _, tx := range p.Transactions {
		id := strings.ToLower(strings.TrimSpace(tx.CoinID))
		if states[id] == nil {
			states[id] = &costState{currency: normalizeCurrency(tx.Currency)}
		}
		s := states[id]
		txCurrency := normalizeCurrency(tx.Currency)
		if s.currency != txCurrency {
			s.mixed = true
		}

		switch tx.Type {
		case "buy":
			s.totalCost += tx.Amount * tx.Price
			s.amount += tx.Amount
		case "sell":
			if s.amount > 0 {
				avgCost := s.totalCost / s.amount
				s.realizedPnL += (tx.Price - avgCost) * tx.Amount
				s.totalCost -= avgCost * tx.Amount
				s.amount -= tx.Amount
				if isEffectivelyZero(s.amount) {
					s.amount = 0
					s.totalCost = 0
				}
			}
		}
		coinIDs[id] = struct{}{}
	}

	for coinID := range coinIDs {
		amount := p.GetHolding(coinID)
		if isEffectivelyZero(amount) {
			continue
		}

		currentPrice, hasPrice := prices[coinID]
		if !hasPrice {
			continue
		}

		s := states[coinID]
		avgCost := 0.0
		totalCost := 0.0
		realized := 0.0
		txCurrency := displayCurrency
		mixed := false
		if s != nil {
			if s.amount > 0 {
				avgCost = s.totalCost / s.amount
				totalCost = avgCost * amount
			}
			realized = s.realizedPnL
			txCurrency = s.currency
			mixed = s.mixed
		}

		currentValue := amount * currentPrice
		unrealized := currentValue - totalCost
		unrealizedPct := 0.0
		if totalCost > 0 {
			unrealizedPct = (unrealized / totalCost) * 100
		}

		if txCurrency != displayCurrency {
			mixed = true
			result.HasMixedCurrency = true
		}

		coinPnL := CoinPnL{
			CoinID:              coinID,
			Amount:              amount,
			AvgCost:             avgCost,
			TotalCost:           totalCost,
			CurrentPrice:        currentPrice,
			CurrentValue:        currentValue,
			UnrealizedPnL:       unrealized,
			UnrealizedPnLPct:    unrealizedPct,
			RealizedPnL:         realized,
			HasMixedCurrency:    mixed,
			TransactionCurrency: txCurrency,
		}
		result.Coins = append(result.Coins, coinPnL)
		result.TotalValue += currentValue
		result.TotalCost += totalCost
		result.TotalUnrealizedPnL += unrealized
		result.TotalRealizedPnL += realized
	}

	return result
}
