package routes

import (
	"encoding/json"
	"net/http"
)

type DashboardResponse struct {
	Stats struct {
		GrossVolume30d int `json:"grossVolume30d"`
		GrossVolume6m  int `json:"grossVolume6m"`
		GrossVolume1y  int `json:"grossVolume1y"`
	} `json:"stats"`

	RecentActivity []Activity `json:"recentActivity"`
}

type Activity struct {
	ID       string `json:"id"`
	Amount   int    `json:"amount"`
	Status   string `json:"status"`
	DateTime string `json:"dateTime"`
}

func (h *Handler) GetDashboardHandler(w http.ResponseWriter, r *http.Request) {

	response := DashboardResponse{}

	// Mock data — now served from backend
	response.Stats.GrossVolume30d = 126420
	response.Stats.GrossVolume6m = 712860
	response.Stats.GrossVolume1y = 1487300

	response.RecentActivity = []Activity{
		{ID: "pay_10021", Amount: 4200, Status: "Settled", DateTime: "2026-03-01 09:58 AM"},
		{ID: "pay_10020", Amount: 870, Status: "Pending", DateTime: "2026-03-01 09:51 AM"},
		{ID: "pay_10019", Amount: 1940, Status: "Settled", DateTime: "2026-03-01 09:39 AM"},
		{ID: "pay_10018", Amount: 2560, Status: "Failed", DateTime: "2026-03-01 09:28 AM"},
		{ID: "pay_10017", Amount: 640, Status: "Settled", DateTime: "2026-03-01 09:07 AM"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
