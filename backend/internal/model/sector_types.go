package model

// ICBSector represents an ICB sector index code
type ICBSector string

const (
	VNIT   ICBSector = "VNIT"   // Công nghệ thông tin (Technology)
	VNIND  ICBSector = "VNIND"  // Công nghiệp (Industrial)
	VNCONS ICBSector = "VNCONS" // Hàng tiêu dùng (Consumer Staples)
	VNCOND ICBSector = "VNCOND" // Hàng tiêu dùng thiết yếu (Consumer Discretionary)
	VNHEAL ICBSector = "VNHEAL" // Chăm sóc sức khỏe (Healthcare)
	VNENE  ICBSector = "VNENE"  // Năng lượng (Energy)
	VNUTI  ICBSector = "VNUTI"  // Tiện ích (Utilities)
	VNREAL ICBSector = "VNREAL" // Bất động sản (Real Estate)
	VNFIN  ICBSector = "VNFIN"  // Tài chính (Finance)
	VNMAT  ICBSector = "VNMAT"  // Nguyên vật liệu (Materials)
)

// AllICBSectors contains all 10 ICB sector indices
var AllICBSectors = []ICBSector{VNIT, VNIND, VNCONS, VNCOND, VNHEAL, VNENE, VNUTI, VNREAL, VNFIN, VNMAT}

// SectorNameMap maps ICB sector codes to Vietnamese names
var SectorNameMap = map[ICBSector]string{
	VNIT:   "Công nghệ thông tin",
	VNIND:  "Công nghiệp",
	VNCONS: "Hàng tiêu dùng",
	VNCOND: "Hàng tiêu dùng thiết yếu",
	VNHEAL: "Chăm sóc sức khỏe",
	VNENE:  "Năng lượng",
	VNUTI:  "Tiện ích",
	VNREAL: "Bất động sản",
	VNFIN:  "Tài chính",
	VNMAT:  "Nguyên vật liệu",
}

// SectorTrend represents the trend direction of a sector
type SectorTrend string

const (
	Uptrend   SectorTrend = "uptrend"
	Downtrend SectorTrend = "downtrend"
	Sideways  SectorTrend = "sideways"
)

// SectorPerformance holds performance metrics for a single ICB sector
type SectorPerformance struct {
	Sector           ICBSector   `json:"sector"`
	SectorName       string      `json:"sectorName"`
	Trend            SectorTrend `json:"trend"`
	TodayChange      float64     `json:"todayChange"`
	OneWeekChange    float64     `json:"oneWeekChange"`
	OneMonthChange   float64     `json:"oneMonthChange"`
	ThreeMonthChange float64     `json:"threeMonthChange"`
	SixMonthChange   float64     `json:"sixMonthChange"`
	OneYearChange    float64     `json:"oneYearChange"`
	CurrentPrice     float64     `json:"currentPrice"`
	SMA20            float64     `json:"sma20"`
	SMA50            float64     `json:"sma50"`
	IsStale          bool        `json:"isStale"`
}

// SectorAverages holds median fundamental metrics for stocks in a sector
type SectorAverages struct {
	Sector             ICBSector `json:"sector"`
	MedianPE           float64   `json:"medianPE"`
	MedianPB           float64   `json:"medianPB"`
	MedianROE          float64   `json:"medianROE"`
	MedianROA          float64   `json:"medianROA"`
	MedianDivYield     float64   `json:"medianDivYield"`
	MedianDebtToEquity float64   `json:"medianDebtToEquity"`
}
