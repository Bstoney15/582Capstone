package routes

import (
	"net/http"
	"os"
	"path/filepath"
)

func (h *Handler) GetCheckoutWidget(w http.ResponseWriter, r *http.Request) {

	widgetPath := filepath.Clean("./static/widget/checkout.js")

	if _, err := os.Stat(widgetPath); err != nil {
		http.Error(w, "checkout widget not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	http.ServeFile(w, r, widgetPath)
}
