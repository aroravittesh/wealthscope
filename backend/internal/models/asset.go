package models

import "time"

type Asset struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	AssetType string    `json:"asset_type"`
	CreatedAt time.Time `json:"created_at"`
}
