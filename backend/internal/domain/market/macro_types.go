package market

import "time"

// MacroIndicators contains aggregated Vietnamese macroeconomic data.
type MacroIndicators struct {
	InterbankRates []InterbankRate  `json:"interbankRates"`
	BondYields     []BondYield      `json:"bondYields"`
	FXRates        []MacroFXRate    `json:"fxRates,omitempty"`
	GoldPrices     []MacroGoldPrice `json:"goldPrices,omitempty"`
	WorldIndices   []WorldIndex     `json:"worldIndices,omitempty"`
	CPI            *float64         `json:"cpi,omitempty"`
	GDPGrowth      *float64         `json:"gdpGrowth,omitempty"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

type InterbankRate struct {
	Tenor string  `json:"tenor"`
	Rate  float64 `json:"rate"`
}

type BondYield struct {
	Tenor string  `json:"tenor"`
	Yield float64 `json:"yield"`
}

// MacroFXRate represents a foreign exchange rate (renamed to avoid clash with DataCategory constant).
type MacroFXRate struct {
	Pair   string  `json:"pair"`
	Rate   float64 `json:"rate"`
	Source string  `json:"source"`
}

// MacroGoldPrice represents a gold price entry (renamed to avoid clash with DataCategory constant).
type MacroGoldPrice struct {
	TypeName  string  `json:"typeName"`
	BuyPrice  float64 `json:"buyPrice"`
	SellPrice float64 `json:"sellPrice"`
	Source    string  `json:"source"`
}

type WorldIndex struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Value         float64 `json:"value"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
}
