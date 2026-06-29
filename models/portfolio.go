/*
Copyright © 2024 Serkan MERCAN <serkanmercan@email.com>
*/

package models

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Transaction struct {
	CoinID   string    `json:"coin_id"`
	Symbol   string    `json:"symbol"`
	Amount   float64   `json:"amount"`
	Price    float64   `json:"price"`
	Currency string    `json:"currency,omitempty"`
	Type     string    `json:"type"` // "buy" or "sell"
	Date     time.Time `json:"date"`
}

type Portfolio struct {
	Holdings     map[string]float64 `json:"holdings"`
	Transactions []Transaction      `json:"transactions"`
	FilePath     string             `json:"-"`
}

func NewPortfolio(filePath string) *Portfolio {
	return &Portfolio{
		Holdings:     make(map[string]float64),
		Transactions: make([]Transaction, 0),
		FilePath:     filePath,
	}
}

func normalizeTransaction(t *Transaction) {
	t.CoinID = strings.ToLower(strings.TrimSpace(t.CoinID))
	t.Type = strings.ToLower(strings.TrimSpace(t.Type))
	if t.Currency == "" {
		t.Currency = "usd"
	} else {
		t.Currency = strings.ToLower(strings.TrimSpace(t.Currency))
	}
}

func (p *Portfolio) AddTransaction(t Transaction) error {
	normalizeTransaction(&t)

	if t.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if t.Price <= 0 {
		return fmt.Errorf("price must be greater than zero")
	}
	if t.Type != "buy" && t.Type != "sell" {
		return fmt.Errorf("transaction type must be 'buy' or 'sell'")
	}

	t.Date = time.Now()

	if t.Type == "buy" {
		p.Holdings[t.CoinID] += t.Amount
	} else {
		if p.Holdings[t.CoinID] < t.Amount {
			return fmt.Errorf("insufficient balance")
		}
		p.Holdings[t.CoinID] -= t.Amount
		if p.Holdings[t.CoinID] == 0 {
			delete(p.Holdings, t.CoinID)
		}
	}

	p.Transactions = append(p.Transactions, t)
	return p.Save()
}

func (p *Portfolio) GetHolding(coinID string) float64 {
	return p.Holdings[strings.ToLower(strings.TrimSpace(coinID))]
}

func (p *Portfolio) Save() error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.FilePath, data, privateFileMode)
}

func (p *Portfolio) Load() error {
	data, err := os.ReadFile(p.FilePath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}
	p.normalizeLoadedData()
	return nil
}

func (p *Portfolio) normalizeLoadedData() {
	normalizedHoldings := make(map[string]float64)
	for coinID, amount := range p.Holdings {
		id := strings.ToLower(strings.TrimSpace(coinID))
		normalizedHoldings[id] += amount
	}
	for coinID, amount := range normalizedHoldings {
		if amount == 0 {
			delete(normalizedHoldings, coinID)
		}
	}
	p.Holdings = normalizedHoldings

	for i := range p.Transactions {
		normalizeTransaction(&p.Transactions[i])
	}
}

func (p *Portfolio) CalculateValue(prices map[string]float64) float64 {
	totalValue := 0.0
	for coinID, amount := range p.Holdings {
		if price, ok := prices[coinID]; ok {
			totalValue += amount * price
		}
	}
	return totalValue
}
