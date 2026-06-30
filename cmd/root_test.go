package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mrcnserkan/crypto/v2/models"
	"github.com/spf13/cobra"
)

func setupTestEnv(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	configDir = dir
	portfolio = models.NewPortfolio(filepath.Join(dir, "portfolio.json"))
	alertManager = models.NewAlertManager(dir)
	configStore = models.NewConfigStore(dir)
	watchlist = models.NewWatchlist(dir)
	daemonState = models.NewDaemonState(dir)
}

func TestGetCurrencyFlagUsesConfigDefault(t *testing.T) {
	setupTestEnv(t)
	configStore.Config.Currency = "eur"

	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("currency", "", "currency")
	if got := getCurrencyFlag(cmd); got != "eur" {
		t.Fatalf("getCurrencyFlag() = %q, want eur", got)
	}
}

func TestAlertListEmpty(t *testing.T) {
	setupTestEnv(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	alertListCmd.Run(alertListCmd, []string{})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	if !bytes.Contains(buf.Bytes(), []byte("No active alerts")) {
		t.Fatalf("expected empty alert message, got: %s", buf.String())
	}
}

func TestVersionConstant(t *testing.T) {
	if Version != "v2.0.1" {
		t.Fatalf("Version = %q, want v2.0.1", Version)
	}
}

func TestPortfolioHasAnyDataModel(t *testing.T) {
	setupTestEnv(t)
	_ = portfolio.AddTransaction(models.Transaction{
		CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 100, Currency: "usd", Type: "buy",
	})
	_ = portfolio.AddTransaction(models.Transaction{
		CoinID: "bitcoin", Symbol: "btc", Amount: 1, Price: 120, Currency: "usd", Type: "sell",
	})
	if portfolio.HasHoldings() {
		t.Fatal("expected no holdings")
	}
	if !portfolio.HasAnyData() {
		t.Fatal("expected history-only data")
	}
}
