package dto_request

type AdminDashboardReportRequest struct {
	StartDate string `form:"start_date"` // format: YYYY-MM-DD
	EndDate   string `form:"end_date"`   // format: YYYY-MM-DD
}
