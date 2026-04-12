// Authors: Ben Stonestreet, Joe Hotze, Ryan Grimsley
// Created: 03/4/26
// Description: backend file containing handler functions for the merchant dashboard
package routes

import (
	"backend/libraries/sessionManager"
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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

func queryDashboardGrossVolumeStats(db *gorm.DB, merchantID string, now time.Time) (DashboardGrossVolumeStats, error) {
	type totalResult struct {
		Total decimal.Decimal `gorm:"column:total"`
	}

	sumSince := func(since time.Time) (int, error) {
		result := totalResult{Total: decimal.Zero}
		err := db.Model(&models.Invoice{}).
			Joins("JOIN merchant_customers ON merchant_customers.customer_id = invoice.invoice_customer_id").
			Select("COALESCE(SUM(invoice_amount_charged), 0) AS total").
			Where("merchant_customers.customer_merchant_id = ?", merchantID).
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

func queryDashboardRecentActivity(db *gorm.DB, merchantID string, limit int) ([]DashboardInvoiceActivity, error) {
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
		Joins("JOIN merchant_customers ON merchant_customers.customer_id = invoice.invoice_customer_id").
		Select("invoice_id, invoice_amount_charged, invoice_status, invoice_date_time").
		Where("merchant_customers.customer_merchant_id = ?", merchantID).
		Where("invoice_status = ?", "paid").
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

func queryDashboardSearchInvoices(db *gorm.DB, merchantID string, invoiceIDSearch string) ([]DashboardInvoiceActivity, error) {
	type activityRow struct {
		InvoiceID            string          `gorm:"column:invoice_id"`
		InvoiceAmountCharged decimal.Decimal `gorm:"column:invoice_amount_charged"`
		InvoiceStatus        string          `gorm:"column:invoice_status"`
		InvoiceDateTime      time.Time       `gorm:"column:invoice_date_time"`
	}

	rows := []activityRow{}
	err := db.Model(&models.Invoice{}).
		Joins("JOIN merchant_customers ON merchant_customers.customer_id = invoice.invoice_customer_id").
		Select("invoice_id, invoice_amount_charged, invoice_status, invoice_date_time").
		Where("merchant_customers.customer_merchant_id = ?", merchantID).
		Where("invoice_id LIKE ?", "%"+invoiceIDSearch+"%").
		Order("invoice_date_time DESC").
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
	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionData, _, active := sessionManager.CheckSession(sessionToken)
	if !active {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	merchantID := strings.TrimSpace(r.URL.Query().Get("merchant_id"))
	if merchantID == "" {
		http.Error(w, "merchant_id is required", http.StatusBadRequest)
		return
	}

	var role models.Role
	if err := h.DB.Where(
		"role_user_id = ? AND role_merchant_id = ? AND role_name IN ?",
		sessionData.UserID,
		merchantID,
		[]string{models.RoleDeveloper, models.RoleAdmin, models.RoleOwner},
	).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		http.Error(w, "failed to verify merchant access", http.StatusInternalServerError)
		return
	}

	stats, err := queryDashboardGrossVolumeStats(h.DB, merchantID, time.Now())
	if err != nil {
		http.Error(w, "failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}
	var recentActivityLimit int = 20
	recentActivity, err := queryDashboardRecentActivity(h.DB, merchantID, recentActivityLimit)
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

func (h *Handler) SearchInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := sessionManager.GetSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionData, _, active := sessionManager.CheckSession(sessionToken)
	if !active {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	merchantID := strings.TrimSpace(r.URL.Query().Get("merchant_id"))
	if merchantID == "" {
		http.Error(w, "merchant_id is required", http.StatusBadRequest)
		return
	}

	invoiceIDSearch := strings.TrimSpace(r.URL.Query().Get("invoice_id"))
	if invoiceIDSearch == "" {
		http.Error(w, "invoice_id search term is required", http.StatusBadRequest)
		return
	}

	var role models.Role
	if err := h.DB.Where(
		"role_user_id = ? AND role_merchant_id = ? AND role_name IN ?",
		sessionData.UserID,
		merchantID,
		[]string{models.RoleDeveloper, models.RoleAdmin, models.RoleOwner},
	).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		http.Error(w, "failed to verify merchant access", http.StatusInternalServerError)
		return
	}

	searchResults, err := queryDashboardSearchInvoices(h.DB, merchantID, invoiceIDSearch)
	if err != nil {
		http.Error(w, "failed to search invoices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchResults)
}
