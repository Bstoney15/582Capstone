package routes

import (
	"backend/libraries/apiauth"
	"backend/models"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type createCustomerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type updateCustomerRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
}

type createInvoiceMerchantRequest struct {
	CustomerID    string `json:"customer_id"`
	AmountCharged string `json:"amount_charged"`
	Status        string `json:"status"`
	FeeAmount     string `json:"fee_amount"`
	FeeStatus     string `json:"fee_status"`
	CryptoType    string `json:"crypto_type"`
}

type updateInvoiceMerchantRequest struct {
	CustomerID    *string `json:"customer_id"`
	AmountCharged *string `json:"amount_charged"`
	Status        *string `json:"status"`
	FeeAmount     *string `json:"fee_amount"`
	FeeStatus     *string `json:"fee_status"`
	CryptoType    *string `json:"crypto_type"`
}

func requireMerchantIDFromAPIKey(w http.ResponseWriter, r *http.Request) (string, bool) {
	merchantID, ok := apiauth.MerchantIDFromContext(r.Context())
	if !ok {
		http.Error(w, "invalid merchant api key", http.StatusUnauthorized)
		return "", false
	}
	return merchantID, true
}

func (h *Handler) CreateMerchantCustomerHandler(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := requireMerchantIDFromAPIKey(w, r)
	if !ok {
		return
	}

	var request createCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	customer := models.Customer{
		CustomerID:         uuid.New().String(),
		CustomerMerchantID: merchantID,
		CustomerFirstName:  strings.TrimSpace(request.FirstName),
		CustomerLastName:   strings.TrimSpace(request.LastName),
		CustomerEmail:      strings.TrimSpace(request.Email),
	}

	if err := h.DB.Create(&customer).Error; err != nil {
		http.Error(w, "failed to create customer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(customer)
}

func (h *Handler) ListMerchantCustomersHandler(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := requireMerchantIDFromAPIKey(w, r)
	if !ok {
		return
	}

	var customers []models.Customer
	if err := h.DB.Where("customer_merchant_id = ?", merchantID).Order("customer_id DESC").Find(&customers).Error; err != nil {
		http.Error(w, "failed to fetch customers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

func (h *Handler) GetMerchantCustomerHandler(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := requireMerchantIDFromAPIKey(w, r)
	if !ok {
		return
	}

	customerID := strings.TrimSpace(r.PathValue("customer_id"))
	if customerID == "" {
		http.Error(w, "customer_id is required", http.StatusBadRequest)
		return
	}

	var customer models.Customer
	err := h.DB.Where("customer_id = ? AND customer_merchant_id = ?", customerID, merchantID).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "customer not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch customer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

func (h *Handler) UpdateMerchantCustomerHandler(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := requireMerchantIDFromAPIKey(w, r)
	if !ok {
		return
	}

	customerID := strings.TrimSpace(r.PathValue("customer_id"))
	if customerID == "" {
		http.Error(w, "customer_id is required", http.StatusBadRequest)
		return
	}

	var request updateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updates := map[string]interface{}{}
	if request.FirstName != nil {
		updates["customer_first_name"] = strings.TrimSpace(*request.FirstName)
	}
	if request.LastName != nil {
		updates["customer_last_name"] = strings.TrimSpace(*request.LastName)
	}
	if request.Email != nil {
		updates["customer_email"] = strings.TrimSpace(*request.Email)
	}

	if len(updates) == 0 {
		http.Error(w, "no fields provided for update", http.StatusBadRequest)
		return
	}

	result := h.DB.Model(&models.Customer{}).Where("customer_id = ? AND customer_merchant_id = ?", customerID, merchantID).Updates(updates)
	if result.Error != nil {
		http.Error(w, "failed to update customer", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, "customer not found", http.StatusNotFound)
		return
	}

	h.GetMerchantCustomerHandler(w, r)
}

func (h *Handler) DeleteMerchantCustomerHandler(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := requireMerchantIDFromAPIKey(w, r)
	if !ok {
		return
	}

	customerID := strings.TrimSpace(r.PathValue("customer_id"))
	if customerID == "" {
		http.Error(w, "customer_id is required", http.StatusBadRequest)
		return
	}

	result := h.DB.Where("customer_id = ? AND customer_merchant_id = ?", customerID, merchantID).Delete(&models.Customer{})
	if result.Error != nil {
		http.Error(w, "failed to delete customer", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, "customer not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseBoolQuery(value string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return false
	}

	parsed, err := strconv.ParseBool(trimmed)
	if err == nil {
		return parsed
	}

	return trimmed == "1" || trimmed == "yes" || trimmed == "y"
}

func (h *Handler) CreateMerchantInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := requireMerchantIDFromAPIKey(w, r)
	if !ok {
		return
	}

	var request createInvoiceMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	customerID := strings.TrimSpace(request.CustomerID)
	if customerID == "" {
		http.Error(w, "customer_id is required", http.StatusBadRequest)
		return
	}

	var customer models.Customer
	if err := h.DB.Where("customer_id = ? AND customer_merchant_id = ?", customerID, merchantID).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "customer not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to resolve customer", http.StatusInternalServerError)
		return
	}

	amountCharged, err := decimal.NewFromString(strings.TrimSpace(request.AmountCharged))
	if err != nil || !amountCharged.IsPositive() {
		http.Error(w, "amount_charged must be a positive decimal string", http.StatusBadRequest)
		return
	}

	feeAmount := decimal.Zero
	if strings.TrimSpace(request.FeeAmount) != "" {
		parsedFee, feeErr := decimal.NewFromString(strings.TrimSpace(request.FeeAmount))
		if feeErr != nil || parsedFee.IsNegative() {
			http.Error(w, "fee_amount must be a non-negative decimal string", http.StatusBadRequest)
			return
		}
		feeAmount = parsedFee
	}

	status := strings.TrimSpace(request.Status)
	if status == "" {
		status = "created"
	}

	feeStatus := strings.TrimSpace(request.FeeStatus)
	if feeStatus == "" {
		feeStatus = "unpaid"
	}

	cryptoType := strings.TrimSpace(request.CryptoType)
	if cryptoType == "" {
		cryptoType = "XRP"
	}

	invoice := models.Invoice{
		InvoiceID:            uuid.New().String(),
		InvoiceAmountCharged: amountCharged.Round(4),
		InvoiceStatus:        status,
		InvoiceFeeAmount:     feeAmount.Round(4),
		InvoiceFeeStatus:     feeStatus,
		InvoiceCryptoType:    cryptoType,
		InvoiceCustomerID:    customer.CustomerID,
	}

	if err := h.DB.Create(&invoice).Error; err != nil {
		http.Error(w, "failed to create invoice", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(invoice)
}

func (h *Handler) ListMerchantInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := requireMerchantIDFromAPIKey(w, r)
	if !ok {
		return
	}

	query := h.DB.Model(&models.Invoice{}).
		Joins("JOIN merchant_customers ON merchant_customers.customer_id = invoice.invoice_customer_id").
		Where("merchant_customers.customer_merchant_id = ?", merchantID)

	if parseBoolQuery(r.URL.Query().Get("completed_only")) {
		query = query.Where("invoice.invoice_status = ?", "paid")
	}

	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		query = query.Where("invoice.invoice_status = ?", status)
	}

	var invoices []models.Invoice
	if err := query.Order("invoice.invoice_date_time DESC").Find(&invoices).Error; err != nil {
		http.Error(w, "failed to fetch invoices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoices)
}

func (h *Handler) GetMerchantInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := requireMerchantIDFromAPIKey(w, r)
	if !ok {
		return
	}

	invoiceID := strings.TrimSpace(r.PathValue("invoice_id"))
	if invoiceID == "" {
		http.Error(w, "invoice_id is required", http.StatusBadRequest)
		return
	}

	var invoice models.Invoice
	err := h.DB.Model(&models.Invoice{}).
		Joins("JOIN merchant_customers ON merchant_customers.customer_id = invoice.invoice_customer_id").
		Where("merchant_customers.customer_merchant_id = ?", merchantID).
		Where("invoice.invoice_id = ?", invoiceID).
		First(&invoice).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "invoice not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch invoice", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoice)
}

func (h *Handler) UpdateMerchantInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := requireMerchantIDFromAPIKey(w, r)
	if !ok {
		return
	}

	invoiceID := strings.TrimSpace(r.PathValue("invoice_id"))
	if invoiceID == "" {
		http.Error(w, "invoice_id is required", http.StatusBadRequest)
		return
	}

	var request updateInvoiceMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updates := map[string]interface{}{}
	if request.Status != nil {
		updates["invoice_status"] = strings.TrimSpace(*request.Status)
	}
	if request.FeeStatus != nil {
		updates["invoice_fee_status"] = strings.TrimSpace(*request.FeeStatus)
	}
	if request.CryptoType != nil {
		updates["invoice_crypto_type"] = strings.TrimSpace(*request.CryptoType)
	}
	if request.AmountCharged != nil {
		amount, err := decimal.NewFromString(strings.TrimSpace(*request.AmountCharged))
		if err != nil || !amount.IsPositive() {
			http.Error(w, "amount_charged must be a positive decimal string", http.StatusBadRequest)
			return
		}
		updates["invoice_amount_charged"] = amount.Round(4)
	}
	if request.FeeAmount != nil {
		fee, err := decimal.NewFromString(strings.TrimSpace(*request.FeeAmount))
		if err != nil || fee.IsNegative() {
			http.Error(w, "fee_amount must be a non-negative decimal string", http.StatusBadRequest)
			return
		}
		updates["invoice_fee_amount"] = fee.Round(4)
	}
	if request.CustomerID != nil {
		customerID := strings.TrimSpace(*request.CustomerID)
		if customerID == "" {
			http.Error(w, "customer_id cannot be empty", http.StatusBadRequest)
			return
		}

		var customer models.Customer
		if err := h.DB.Where("customer_id = ? AND customer_merchant_id = ?", customerID, merchantID).First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "customer not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to resolve customer", http.StatusInternalServerError)
			return
		}
		updates["invoice_customer_id"] = customerID
	}

	if len(updates) == 0 {
		http.Error(w, "no fields provided for update", http.StatusBadRequest)
		return
	}

	var scopedInvoice models.Invoice
	err := h.DB.Model(&models.Invoice{}).
		Joins("JOIN merchant_customers ON merchant_customers.customer_id = invoice.invoice_customer_id").
		Where("merchant_customers.customer_merchant_id = ?", merchantID).
		Where("invoice.invoice_id = ?", invoiceID).
		First(&scopedInvoice).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "invoice not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch invoice", http.StatusInternalServerError)
		return
	}

	result := h.DB.Model(&scopedInvoice).Updates(updates)
	if result.Error != nil {
		http.Error(w, "failed to update invoice", http.StatusInternalServerError)
		return
	}

	h.GetMerchantInvoiceHandler(w, r)
}

func (h *Handler) DeleteMerchantInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	merchantID, ok := requireMerchantIDFromAPIKey(w, r)
	if !ok {
		return
	}

	invoiceID := strings.TrimSpace(r.PathValue("invoice_id"))
	if invoiceID == "" {
		http.Error(w, "invoice_id is required", http.StatusBadRequest)
		return
	}

	var scopedInvoice models.Invoice
	err := h.DB.Model(&models.Invoice{}).
		Joins("JOIN merchant_customers ON merchant_customers.customer_id = invoice.invoice_customer_id").
		Where("merchant_customers.customer_merchant_id = ?", merchantID).
		Where("invoice.invoice_id = ?", invoiceID).
		First(&scopedInvoice).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "invoice not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch invoice", http.StatusInternalServerError)
		return
	}

	result := h.DB.Delete(&scopedInvoice)
	if result.Error != nil {
		http.Error(w, "failed to delete invoice", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
