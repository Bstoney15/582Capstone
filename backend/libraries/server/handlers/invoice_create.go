package routes

// Author: Benjamin Stonestreet
// Created: 2026-03-08

import (
	"backend/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	defaultFallbackUSDPerXRP = "2.0000"
	xrpUSDQuoteURL           = "https://api.coinbase.com/v2/prices/XRP-USD/spot"
)

// CreateInvoiceRequest is the payload expected for creating a new invoice.
type CreateInvoiceRequest struct {
	MerchantAPIKey string `json:"merchant_api_key"`
	AmountXRP      any    `json:"amount_xrp,omitempty"`
	AmountUSD      any    `json:"amount_usd,omitempty"`
}

// CreateInvoiceResponse is the response returned after a successful invoice creation.
type CreateInvoiceResponse struct {
	InvoiceID     string `json:"invoice_id"`
	AmountXRP     string `json:"amount_xrp"`
	AmountUSD     string `json:"amount_usd,omitempty"`
	USDPerXRP     string `json:"usd_per_xrp,omitempty"`
	PricingSource string `json:"pricing_source,omitempty"`
}

// coinbaseSpotResponse maps the response format from Coinbase's spot price API.
type coinbaseSpotResponse struct {
	Data struct {
		Amount string `json:"amount"`
	} `json:"data"`
}

func parseDecimalInput(value any) (decimal.Decimal, bool, error) {
	if value == nil {
		return decimal.Zero, false, nil
	}

	switch typed := value.(type) {
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return decimal.Zero, false, nil
		}
		parsed, err := decimal.NewFromString(trimmed)
		if err != nil {
			return decimal.Zero, true, err
		}
		return parsed, true, nil
	case float64:
		return decimal.NewFromFloat(typed), true, nil
	case int:
		return decimal.NewFromInt(int64(typed)), true, nil
	case int64:
		return decimal.NewFromInt(typed), true, nil
	case json.Number:
		parsed, err := decimal.NewFromString(typed.String())
		if err != nil {
			return decimal.Zero, true, err
		}
		return parsed, true, nil
	default:
		return decimal.Zero, true, fmt.Errorf("unsupported amount type")
	}
}

// resolveUSDPerXRP fetches the current XRP spot price in USD from Coinbase.
func resolveUSDPerXRP() (decimal.Decimal, string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	request, err := http.NewRequest(http.MethodGet, xrpUSDQuoteURL, nil)
	if err == nil {
		request.Header.Set("Accept", "application/json")
		response, requestErr := client.Do(request)
		if requestErr == nil {
			defer response.Body.Close()
			if response.StatusCode >= 200 && response.StatusCode < 300 {
				var body coinbaseSpotResponse
				if decodeErr := json.NewDecoder(response.Body).Decode(&body); decodeErr == nil {
					rate, parseErr := decimal.NewFromString(body.Data.Amount)
					if parseErr == nil && rate.IsPositive() {
						return rate, "coinbase_spot", nil
					}
				}
			}
		}
	}

	fallback := os.Getenv("XRPL_FALLBACK_USD_PER_XRP")
	if fallback == "" {
		fallback = defaultFallbackUSDPerXRP
	}

	rate, parseErr := decimal.NewFromString(fallback)
	if parseErr != nil || !rate.IsPositive() {
		return decimal.Zero, "", fmt.Errorf("invalid fallback rate: %s", fallback)
	}

	return rate, "fallback_env", nil
}

// CreateInvoiceHandler handles the HTTP route for generating a retail invoice.
func (h *Handler) CreateInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	applyWidgetCORSHeaders(w, r)

	var request CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	merchantAPIKey := strings.TrimSpace(request.MerchantAPIKey)
	if merchantAPIKey == "" {
		http.Error(w, "invalid merchant api key", http.StatusUnauthorized)
		return
	}

	merchantAPIKeyHash := hashAPIKey(merchantAPIKey)

	var apiKey models.MerchantAPIKey
	if err := h.DB.Where("merchant_api_key_hashed IN ?", []string{merchantAPIKeyHash, merchantAPIKey}).First(&apiKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "invalid merchant api key", http.StatusUnauthorized)
			return
		}

		http.Error(w, "failed to validate merchant api key", http.StatusInternalServerError)
		return
	}

	if strings.TrimSpace(apiKey.MerchantAPIKeyMerchantID) == "" {
		http.Error(w, "invalid merchant api key", http.StatusUnauthorized)
		return
	}

	var amountXRP decimal.Decimal
	var amountUSD decimal.Decimal
	pricingSource := "direct_xrp"
	usdPerXRP := decimal.Zero

	parsedUSD, hasUSD, parseUSDErr := parseDecimalInput(request.AmountUSD)
	if parseUSDErr != nil {
		http.Error(w, "amount_usd must be a valid decimal", http.StatusBadRequest)
		return
	}

	parsedXRP, hasXRP, parseXRPErr := parseDecimalInput(request.AmountXRP)
	if parseXRPErr != nil {
		http.Error(w, "amount_xrp must be a valid decimal", http.StatusBadRequest)
		return
	}

	if hasUSD {
		if !parsedUSD.IsPositive() {
			http.Error(w, "amount_usd must be a positive decimal string", http.StatusBadRequest)
			return
		}

		rate, source, rateErr := resolveUSDPerXRP()
		if rateErr != nil {
			http.Error(w, "failed to resolve USD/XRP rate", http.StatusServiceUnavailable)
			return
		}

		amountUSD = parsedUSD.Round(2)
		usdPerXRP = rate.Round(6)
		amountXRP = amountUSD.Div(rate).Round(4)
		pricingSource = source
	} else {
		if !hasXRP || !parsedXRP.IsPositive() {
			http.Error(w, "either amount_usd or amount_xrp must be a positive decimal string", http.StatusBadRequest)
			return
		}

		amountXRP = parsedXRP.Round(4)
	}

	var customer models.Customer
	if err := h.DB.Where("customer_merchant_id = ?", apiKey.MerchantAPIKeyMerchantID).Order("customer_id ASC").First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "no customer exists for this merchant", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to resolve customer", http.StatusInternalServerError)
		return
	}

	invoice := models.Invoice{
		InvoiceID:            uuid.New().String(),
		InvoiceAmountCharged: amountXRP,
		InvoiceStatus:        "created",
		InvoiceFeeAmount:     decimal.Zero,
		InvoiceFeeStatus:     "unpaid",
		InvoiceCryptoType:    "XRP",
		InvoiceCustomerID:    customer.CustomerID,
	}

	if err := h.DB.Create(&invoice).Error; err != nil {
		http.Error(w, "failed to create invoice", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := CreateInvoiceResponse{
		InvoiceID:     invoice.InvoiceID,
		AmountXRP:     amountXRP.StringFixed(4),
		PricingSource: pricingSource,
	}

	if amountUSD.IsPositive() {
		response.AmountUSD = amountUSD.StringFixed(2)
		response.USDPerXRP = usdPerXRP.StringFixed(6)
	}

	json.NewEncoder(w).Encode(response)
}
