/*
Copyright © 2024 Serkan MERCAN <serkanmercan@email.com>
*/

package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type Transaction struct {
	CoinID string    `json:"coin_id"`
	Symbol string    `json:"symbol"`
	Amount float64   `json:"amount"`
	Price  float64   `json:"price"`
	Type   string    `json:"type"` // "buy" or "sell"
	Date   time.Time `json:"date"`
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

func (p *Portfolio) AddTransaction(t Transaction) error {
	t.Date = time.Now()

	if t.Type == "buy" {
		p.Holdings[t.CoinID] += t.Amount
	} else if t.Type == "sell" {
		if p.Holdings[t.CoinID] < t.Amount {
			return fmt.Errorf("insufficient balance")
		}
		p.Holdings[t.CoinID] -= t.Amount
	}

	p.Transactions = append(p.Transactions, t)
	return p.Save()
}

func (p *Portfolio) GetHolding(coinID string) float64 {
	return p.Holdings[coinID]
}

func (p *Portfolio) Save() error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p.FilePath, data, 0644)
}

func (p *Portfolio) Load() error {
	data, err := ioutil.ReadFile(p.FilePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, p)
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
