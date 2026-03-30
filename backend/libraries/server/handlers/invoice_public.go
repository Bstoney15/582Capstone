package routes

// Author: Benjamin Stonestreet
// Created: 2026-03-08


import (
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// InvoiceCheckoutResponse represents the data returned for public checkout viewing.
type InvoiceCheckoutResponse struct {
	InvoiceID       string `json:"invoiceId"`
	AmountDrops     string `json:"amountDrops"`
	DestinationTag  uint32 `json:"destinationTag"`
	MerchantAddress string `json:"merchantAddress"`
}

// invoiceAmountToDrops converts a decimal XRP amount into stringified drops.
func invoiceAmountToDrops(amount decimal.Decimal) string {
	drops := amount.Mul(decimal.NewFromInt(1_000_000)).Round(0)
	if drops.IsNegative() {
		return "0"
	}

	return drops.StringFixed(0)
}

// GetInvoiceForCheckoutHandler retrieves an invoice for the public facing checkout widget.
func (h *Handler) GetInvoiceForCheckoutHandler(w http.ResponseWriter, r *http.Request) {
	applyWidgetCORSHeaders(w, r)

	invoiceID := r.PathValue("uuid")
	if invoiceID == "" {
		http.Error(w, "invoice id is required", http.StatusBadRequest)
		return
	}

	var invoice models.Invoice
	if err := h.DB.Where("invoice_id = ?", invoiceID).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "invoice not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to fetch invoice", http.StatusInternalServerError)
		return
	}

	var customer models.Customer
	if err := h.DB.Where("customer_id = ?", invoice.InvoiceCustomerID).First(&customer).Error; err != nil {
		http.Error(w, "failed to resolve customer", http.StatusInternalServerError)
		return
	}

	var wallet models.MerchantCryptoWallet
	if err := h.DB.Where(
		"merchant_crypto_wallet_merchant_id = ? AND merchant_crypto_wallet_verified = ?",
		customer.CustomerMerchantID,
		true,
	).First(&wallet).Error; err != nil {
		http.Error(w, "merchant wallet unavailable", http.StatusNotFound)
		return
	}

	response := InvoiceCheckoutResponse{
		InvoiceID:       invoice.InvoiceID,
		AmountDrops:     invoiceAmountToDrops(invoice.InvoiceAmountCharged),
		DestinationTag:  0,
		MerchantAddress: wallet.MerchantCryptoWalletAddress,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
