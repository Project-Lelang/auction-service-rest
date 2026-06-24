package dto_response

type DashboardDailyReport struct {
	Date          string  `json:"date" db:"date"`
	TotalAuctions int     `json:"total_auctions" db:"total_auctions"`
	TotalRevenue  float64 `json:"total_revenue" db:"total_revenue"`
}
