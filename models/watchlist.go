package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Watchlist struct {
	CoinIDs  []string `json:"coin_ids"`
	filePath string
}

func NewWatchlist(configDir string) *Watchlist {
	return &Watchlist{
		CoinIDs:  make([]string, 0),
		filePath: filepath.Join(configDir, "watchlist.json"),
	}
}

func (w *Watchlist) Load() error {
	data, err := os.ReadFile(w.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, w)
}

func (w *Watchlist) Save() error {
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(w.filePath, data)
}

func (w *Watchlist) Add(coinID string) error {
	coinID = strings.ToLower(strings.TrimSpace(coinID))
	if coinID == "" {
		return fmt.Errorf("coin id is required")
	}
	for _, id := range w.CoinIDs {
		if id == coinID {
			return fmt.Errorf("%s is already in watchlist", coinID)
		}
	}
	w.CoinIDs = append(w.CoinIDs, coinID)
	return w.Save()
}

func (w *Watchlist) Remove(coinID string) error {
	coinID = strings.ToLower(strings.TrimSpace(coinID))
	for i, id := range w.CoinIDs {
		if id == coinID {
			w.CoinIDs = append(w.CoinIDs[:i], w.CoinIDs[i+1:]...)
			return w.Save()
		}
	}
	return fmt.Errorf("%s is not in watchlist", coinID)
}

func (w *Watchlist) Contains(coinID string) bool {
	coinID = strings.ToLower(strings.TrimSpace(coinID))
	for _, id := range w.CoinIDs {
		if id == coinID {
			return true
		}
	}
	return false
}
