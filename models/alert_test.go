package models

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAlertManager_AddAndRemoveAllForCoin(t *testing.T) {
	dir := t.TempDir()
	manager := NewAlertManager(dir)

	alerts := []Alert{
		{CoinID: "bitcoin", Price: 50000, Condition: "above", Currency: "usd"},
		{CoinID: "bitcoin", Price: 45000, Condition: "below", Currency: "usd"},
		{CoinID: "ethereum", Price: 2000, Condition: "above", Currency: "usd"},
	}

	for _, alert := range alerts {
		if err := manager.AddAlert(alert); err != nil {
			t.Fatalf("AddAlert() error = %v", err)
		}
	}

	if err := manager.RemoveAlert("bitcoin"); err != nil {
		t.Fatalf("RemoveAlert() error = %v", err)
	}

	remaining := manager.GetAlerts()
	if len(remaining) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(remaining))
	}
	if remaining[0].CoinID != "ethereum" {
		t.Fatalf("expected ethereum alert to remain, got %s", remaining[0].CoinID)
	}
}

func TestAlertManager_RemoveTriggeredAlert(t *testing.T) {
	dir := t.TempDir()
	manager := NewAlertManager(dir)

	first := Alert{CoinID: "bitcoin", Price: 50000, Condition: "above", Currency: "usd"}
	second := Alert{CoinID: "bitcoin", Price: 45000, Condition: "below", Currency: "usd"}

	if err := manager.AddAlert(first); err != nil {
		t.Fatalf("AddAlert(first) error = %v", err)
	}
	if err := manager.AddAlert(second); err != nil {
		t.Fatalf("AddAlert(second) error = %v", err)
	}

	if err := manager.RemoveTriggeredAlert(first); err != nil {
		t.Fatalf("RemoveTriggeredAlert() error = %v", err)
	}

	remaining := manager.GetAlerts()
	if len(remaining) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(remaining))
	}
	if remaining[0].Price != 45000 {
		t.Fatalf("expected remaining alert price 45000, got %.2f", remaining[0].Price)
	}
}

func TestAlertManager_Validation(t *testing.T) {
	dir := t.TempDir()
	manager := NewAlertManager(dir)

	err := manager.AddAlert(Alert{CoinID: "bitcoin", Price: -1, Condition: "above", Currency: "usd"})
	if err == nil {
		t.Fatal("expected error for negative price")
	}

	err = manager.AddAlert(Alert{CoinID: "bitcoin", Price: 100, Condition: "sideways", Currency: "usd"})
	if err == nil {
		t.Fatal("expected error for invalid condition")
	}
}

func TestAlertManager_LoadNormalizesCoinIDAndCurrency(t *testing.T) {
	dir := t.TempDir()
	alertFile := filepath.Join(dir, "alerts.json")
	payload := `[{"coin_id":"BITCOIN","price":100,"condition":"ABOVE","created_at":"2024-01-01T00:00:00Z"}]`
	if err := os.WriteFile(alertFile, []byte(payload), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	manager := NewAlertManager(dir)
	if err := manager.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	alerts := manager.GetAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].CoinID != "bitcoin" || alerts[0].Condition != "above" || alerts[0].Currency != "usd" {
		t.Fatalf("unexpected normalized alert: %+v", alerts[0])
	}
}
