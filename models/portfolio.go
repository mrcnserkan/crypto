/*
Copyright © 2024 Serkan MERCAN <serkanmercan@email.com>
*/

package models

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

const holdingDustThreshold = 1e-10

func isEffectivelyZero(value float64) bool {
	return math.Abs(value) < holdingDustThreshold
}

func normalizeCurrency(currency string) string {
	currency = strings.ToLower(strings.TrimSpace(currency))
	if currency == "" {
		return "usd"
	}
	return currency
}

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
	t.Currency = normalizeCurrency(t.Currency)
}

func (p *Portfolio) pruneDustHoldings() {
	for coinID, amount := range p.Holdings {
		if isEffectivelyZero(amount) {
			delete(p.Holdings, coinID)
		}
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
		holding := p.Holdings[t.CoinID]
		if t.Amount > holding+holdingDustThreshold {
			return fmt.Errorf("insufficient balance")
		}
		p.Holdings[t.CoinID] = holding - t.Amount
		if isEffectivelyZero(p.Holdings[t.CoinID]) {
			delete(p.Holdings, t.CoinID)
		}
	}

	p.pruneDustHoldings()
	p.Transactions = append(p.Transactions, t)
	return p.Save()
}

func (p *Portfolio) GetHolding(coinID string) float64 {
	return p.Holdings[strings.ToLower(strings.TrimSpace(coinID))]
}

func (p *Portfolio) HasHoldings() bool {
	p.pruneDustHoldings()
	return len(p.Holdings) > 0
}

// HasAnyData returns true if portfolio has holdings or transaction history.
func (p *Portfolio) HasAnyData() bool {
	p.pruneDustHoldings()
	return len(p.Holdings) > 0 || len(p.Transactions) > 0
}

// HasCoinData returns true if the coin has holdings or transaction history.
func (p *Portfolio) HasCoinData(coinID string) bool {
	coinID = strings.ToLower(strings.TrimSpace(coinID))
	if amount, ok := p.Holdings[coinID]; ok && !isEffectivelyZero(amount) {
		return true
	}
	for _, t := range p.Transactions {
		if t.CoinID == coinID {
			return true
		}
	}
	return false
}

// Clear removes all holdings and transactions.
func (p *Portfolio) Clear() error {
	p.Holdings = make(map[string]float64)
	p.Transactions = []Transaction{}
	return p.Save()
}

// RemoveCoin removes a coin from holdings and its transaction history.
func (p *Portfolio) RemoveCoin(coinID string) error {
	coinID = strings.ToLower(strings.TrimSpace(coinID))
	delete(p.Holdings, coinID)
	filtered := make([]Transaction, 0, len(p.Transactions))
	for _, t := range p.Transactions {
		if t.CoinID != coinID {
			filtered = append(filtered, t)
		}
	}
	p.Transactions = filtered
	return p.Save()
}

func (p *Portfolio) Save() error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(p.FilePath, data)
}

func (p *Portfolio) Load() error {
	data, err := os.ReadFile(p.FilePath)
	if err != nil {
		return err
	}
	before := string(data)
	if err := json.Unmarshal(data, p); err != nil {
		return err
	}
	p.normalizeLoadedData()
	after, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if string(after) != before {
		return p.Save()
	}
	return nil
}

func (p *Portfolio) normalizeLoadedData() {
	normalizedHoldings := make(map[string]float64)
	for coinID, amount := range p.Holdings {
		id := strings.ToLower(strings.TrimSpace(coinID))
		normalizedHoldings[id] += amount
	}
	p.Holdings = normalizedHoldings
	p.pruneDustHoldings()

	for i := range p.Transactions {
		normalizeTransaction(&p.Transactions[i])
	}
}
