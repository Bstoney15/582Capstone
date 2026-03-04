package routes

import (
	"backend/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DashboardResponse struct {
	Stats          DashboardGrossVolumeStats  `json:"stats"`
	RecentActivity []DashboardInvoiceActivity `json:"recentActivity"`
}

type DashboardGrossVolumeStats struct {
	GrossVolume30d int `json:"grossVolume30d"`
	GrossVolume6m  int `json:"grossVolume6m"`
	GrossVolume1y  int `json:"grossVolume1y"`
}

type DashboardInvoiceActivity struct {
	ID       string `json:"id"`
	Amount   int    `json:"amount"`
	Status   string `json:"status"`
	DateTime string `json:"dateTime"`
}

func queryDashboardGrossVolumeStats(db *gorm.DB, now time.Time) (DashboardGrossVolumeStats, error) {
	type totalResult struct {
		Total decimal.Decimal `gorm:"column:total"`
	}

	sumSince := func(since time.Time) (int, error) {
		result := totalResult{Total: decimal.Zero}
		err := db.Model(&models.Invoice{}).
			Select("COALESCE(SUM(invoice_amount_charged), 0) AS total").
			Where("invoice_date_time >= ?", since).
			Scan(&result).Error
		if err != nil {
			return 0, err
		}

		return int(result.Total.IntPart()), nil
	}

	stats := DashboardGrossVolumeStats{}
	var err error

	stats.GrossVolume30d, err = sumSince(now.AddDate(0, 0, -30))
	if err != nil {
		return DashboardGrossVolumeStats{}, err
	}

	stats.GrossVolume6m, err = sumSince(now.AddDate(0, -6, 0))
	if err != nil {
		return DashboardGrossVolumeStats{}, err
	}

	stats.GrossVolume1y, err = sumSince(now.AddDate(-1, 0, 0))
	if err != nil {
		return DashboardGrossVolumeStats{}, err
	}

	return stats, nil
}

func queryDashboardRecentActivity(db *gorm.DB, limit int) ([]DashboardInvoiceActivity, error) {
	if limit <= 0 {
		limit = 5
	}

	type activityRow struct {
		InvoiceID            string          `gorm:"column:invoice_id"`
		InvoiceAmountCharged decimal.Decimal `gorm:"column:invoice_amount_charged"`
		InvoiceStatus        string          `gorm:"column:invoice_status"`
		InvoiceDateTime      time.Time       `gorm:"column:invoice_date_time"`
	}

	rows := []activityRow{}
	err := db.Model(&models.Invoice{}).
		Select("invoice_id, invoice_amount_charged, invoice_status, invoice_date_time").
		Order("invoice_date_time DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	activities := make([]DashboardInvoiceActivity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, DashboardInvoiceActivity{
			ID:       row.InvoiceID,
			Amount:   int(row.InvoiceAmountCharged.IntPart()),
			Status:   mapInvoiceStatusForDashboard(row.InvoiceStatus),
			DateTime: row.InvoiceDateTime.Local().Format("2006-01-02 03:04 PM"),
		})
	}

	return activities, nil
}

func mapInvoiceStatusForDashboard(status string) string {
	switch status {
	case "paid":
		return "Settled"
	case "created", "verification_pending":
		return "Pending"
	case "verification_failed":
		return "Failed"
	default:
		return status
	}
}

func (h *Handler) GetDashboardHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := queryDashboardGrossVolumeStats(h.DB, time.Now())
	if err != nil {
		http.Error(w, "failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}

	recentActivity, err := queryDashboardRecentActivity(h.DB, 5)
	if err != nil {
		http.Error(w, "failed to fetch dashboard activity", http.StatusInternalServerError)
		return
	}

	response := DashboardResponse{
		Stats:          stats,
		RecentActivity: recentActivity,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
