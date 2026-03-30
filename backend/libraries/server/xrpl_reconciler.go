// Author: Benjamin Stonestreet
// Created: 2026-03-03
// Description:
// Package server – this file implements the XRPLReconciler, a background
// worker that periodically polls the XRP Ledger (via JSON-RPC) for incoming
// payments to every verified merchant wallet stored in the database. When a
// matching payment is found it is recorded as an XRPLPayment and the
// corresponding open Invoice is marked as paid.


package server

import (
	"backend/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// defaultXRPLRPCURL points at the Ripple Testnet. Override with the
	// XRPL_RPC_URL environment variable for Mainnet or a private node.
	defaultXRPLRPCURL = "https://s.altnet.rippletest.net:51234"
	// defaultReconcileInterval is how often the reconciler polls the ledger
	// when no explicit interval is configured.
	defaultReconcileInterval = 5 * time.Second

	// Invoice / payment status strings shared between the reconciler and the
	// invoice handlers so that invoice state transitions are consistent.
	xrplPaymentStatusCreated = "created"
	xrplPaymentStatusPending = "verification_pending"
	xrplPaymentStatusPaid    = "paid"

	// xrplAccountTxPageSizeLimit caps the number of transactions returned per
	// RPC page request. The XRPL node may return fewer; pagination via Marker
	// handles the rest.
	xrplAccountTxPageSizeLimit = 200
)

// xrplRPCRequest is the JSON body sent to the XRPL JSON-RPC endpoint.
type xrplRPCRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

// xrplRPCResponse is the top-level wrapper returned by the XRPL node.
type xrplRPCResponse struct {
	Result xrplAccountTxResult `json:"result"`
}

// xrplAccountTxResult holds the paginated transaction list returned by the
// account_tx RPC method, along with pagination metadata.
type xrplAccountTxResult struct {
	Transactions   []xrplTransactionEnvelope `json:"transactions"`
	LedgerIndexMax int64                     `json:"ledger_index_max"`
	// Marker is an opaque cursor returned by the node when more pages exist.
	Marker interface{} `json:"marker"`
	Status string      `json:"status"`
	Error  string      `json:"error"`
}

// xrplTransactionEnvelope wraps a single transaction as returned by account_tx.
// The transaction data appears under "tx" for older API versions and "tx_json"
// for newer ones; the reconciler checks both fields.
type xrplTransactionEnvelope struct {
	Validated bool       `json:"validated"`
	Tx        xrplTx     `json:"tx"`
	TxJSON    xrplTx     `json:"tx_json"`
	Meta      xrplTxMeta `json:"meta"`
}

// xrplTxMeta contains the transaction outcome as determined by the ledger.
type xrplTxMeta struct {
	TransactionResult string `json:"TransactionResult"`
}

// xrplTx carries the fields from an XRPL transaction that the reconciler needs.
// Amount is interface{} because XRP amounts are drop strings while IOU amounts
// are objects; the reconciler only handles the XRP (string) case.
type xrplTx struct {
	TransactionType string      `json:"TransactionType"`
	Destination     string      `json:"Destination"`
	Amount          interface{} `json:"Amount"`
	Hash            string      `json:"hash"`
	LedgerIndex     int64       `json:"ledger_index"`
	DestinationTag  *uint32     `json:"DestinationTag"`
}

// XRPLReconciler polls the XRP Ledger on a fixed interval and reconciles
// incoming payments against open invoices stored in the database.
type XRPLReconciler struct {
	db         *gorm.DB
	httpClient *http.Client
	rpcURL     string
	interval   time.Duration
	// isReconciling prevents concurrent reconciliation runs if a single poll
	// takes longer than the configured interval.
	isReconciling bool
}

// NewXRPLReconciler creates a reconciler configured from environment variables:
//   - XRPL_RPC_URL: full URL of the XRPL JSON-RPC endpoint (default: Testnet).
//   - XRPL_RECONCILE_INTERVAL_SEC: polling interval in seconds (default: 5).
func NewXRPLReconciler(db *gorm.DB) *XRPLReconciler {
	rpcURL := os.Getenv("XRPL_RPC_URL")
	if rpcURL == "" {
		rpcURL = defaultXRPLRPCURL
	}

	interval := defaultReconcileInterval
	if configured := os.Getenv("XRPL_RECONCILE_INTERVAL_SEC"); configured != "" {
		if value, err := strconv.Atoi(configured); err == nil && value > 0 {
			interval = time.Duration(value) * time.Second
		}
	}

	return &XRPLReconciler{
		db:         db,
		httpClient: &http.Client{Timeout: 20 * time.Second},
		rpcURL:     rpcURL,
		interval:   interval,
	}
}

// Start launches the reconciler in a background goroutine. It fires once
// immediately, then repeats on the configured interval. It does not block the
// caller.
func (r *XRPLReconciler) Start() {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		// Run once before the first tick so payments are caught at startup.
		r.runOnce()
		for range ticker.C {
			r.runOnce()
		}
	}()
}

// runOnce performs a single reconciliation pass across all verified merchant
// wallets. If a previous run is still in progress the call is a no-op, which
// prevents unbounded goroutine growth when the ledger or DB is slow.
func (r *XRPLReconciler) runOnce() {
	if r.isReconciling {
		return
	}
	r.isReconciling = true
	defer func() { r.isReconciling = false }()

	// Only reconcile wallets that have been verified by the merchant.
	var wallets []models.MerchantCryptoWallet
	if err := r.db.Where("merchant_crypto_wallet_verified = ?", true).Find(&wallets).Error; err != nil {
		log.Printf("xrpl reconciler: failed to fetch wallets: %v", err)
		return
	}

	for _, wallet := range wallets {
		if wallet.MerchantCryptoWalletAddress == "" {
			continue
		}

		if err := r.reconcileWallet(wallet.MerchantCryptoWalletMerchantID, wallet.MerchantCryptoWalletAddress); err != nil {
			log.Printf("xrpl reconciler: wallet %s failed: %v", wallet.MerchantCryptoWalletAddress, err)
		}
	}
}

// reconcileWallet fetches all transactions for the given address that arrived
// after the last stored checkpoint, processes each validated XRP Payment, and
// advances the checkpoint to the highest ledger index seen. Pagination is
// handled by following the Marker returned by the XRPL node until it is nil.
func (r *XRPLReconciler) reconcileWallet(merchantID string, address string) error {
	checkpoint, err := r.getCheckpoint(address)
	if err != nil {
		return err
	}

	marker := interface{}(nil)
	maxSeenLedger := checkpoint.LastLedgerIndex

	for {
		result, rpcErr := r.fetchAccountTx(address, checkpoint.LastLedgerIndex+1, marker)
		if rpcErr != nil {
			// XRPL returns lgrIdxsInvalid when ledger_index_min is ahead of the current validated ledger.
			// This is expected when there are no new ledgers/transactions yet, especially on fresh dev DBs.
			if strings.Contains(rpcErr.Error(), "lgrIdxsInvalid") {
				break
			}
			return rpcErr
		}

		for _, envelope := range result.Transactions {
			// Prefer tx_json (newer API format) but fall back to tx.
			tx := envelope.Tx
			if tx.TransactionType == "" {
				tx = envelope.TxJSON
			}

			// Skip unvalidated transactions; they may still be in-flight.
			if !envelope.Validated {
				continue
			}

			// Only XRP Payment transactions carry a drop amount we can match.
			if tx.TransactionType != "Payment" {
				continue
			}

			// Only process transactions that succeeded on-ledger.
			if envelope.Meta.TransactionResult != "tesSUCCESS" {
				continue
			}

			// Ignore payments sent to a different address (can happen with
			// multi-hop or partial-payment routes).
			if tx.Destination != address {
				continue
			}

			// Amount must be a drop string; IOU amounts are objects and are
			// not supported by the current invoice matching logic.
			amountDrops, ok := tx.Amount.(string)
			if !ok || amountDrops == "" {
				continue
			}

			if tx.Hash == "" {
				continue
			}

			if tx.LedgerIndex > maxSeenLedger {
				maxSeenLedger = tx.LedgerIndex
			}

			if err := r.recordAndMatchPayment(merchantID, address, tx.Hash, amountDrops, tx.DestinationTag, tx.LedgerIndex); err != nil {
				log.Printf("xrpl reconciler: payment match failed (%s): %v", tx.Hash, err)
			}
		}

		if result.LedgerIndexMax > maxSeenLedger {
			maxSeenLedger = result.LedgerIndexMax
		}

		if result.Marker == nil {
			break
		}

		marker = result.Marker
	}

	if maxSeenLedger > checkpoint.LastLedgerIndex {
		if err := r.updateCheckpoint(address, maxSeenLedger); err != nil {
			return err
		}
	}

	return nil
}

// recordAndMatchPayment persists an XRPLPayment record and, if it is new
// (RowsAffected > 0 after the conflict-ignoring insert), tries to match it
// against the oldest open invoice whose amount equals the payment's XRP value.
// On a successful match the invoice status is set to "paid" and the payment
// row is linked to that invoice – both changes happen inside a single
// transaction to avoid partial updates.
func (r *XRPLReconciler) recordAndMatchPayment(merchantID string, destination string, txHash string, amountDrops string, destinationTag *uint32, ledgerIndex int64) error {
	processedAt := time.Now().UTC()
	payment := models.XRPLPayment{
		TxHash:           txHash,
		Destination:      destination,
		AmountDrops:      amountDrops,
		DestinationTag:   destinationTag,
		LedgerIndex:      ledgerIndex,
		ProcessedAt:      processedAt,
		WalletMerchantID: merchantID,
	}

	// Insert the payment; if the tx_hash already exists (duplicate poll) do nothing.
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&payment)
	if result.Error != nil {
		return result.Error
	}

	// RowsAffected == 0 means this transaction was already processed.
	if result.RowsAffected == 0 {
		return nil
	}

	// Convert drops (integer string) to XRP with 4 decimal places for invoice matching.
	dropsDecimal, err := decimal.NewFromString(amountDrops)
	if err != nil {
		return err
	}

	amountXRP := dropsDecimal.Div(decimal.NewFromInt(1_000_000)).Round(4)

	// Find the oldest open invoice for this merchant with the exact XRP amount.
	var invoice models.Invoice
	findErr := r.db.
		Joins("JOIN merchant_customers ON merchant_customers.customer_id = invoice.invoice_customer_id").
		Where("merchant_customers.customer_merchant_id = ?", merchantID).
		Where("invoice.invoice_status IN ?", []string{xrplPaymentStatusCreated, xrplPaymentStatusPending}).
		Where("invoice.invoice_amount_charged = ?", amountXRP).
		Order("invoice.invoice_id ASC").
		First(&invoice).Error

	if findErr != nil {
		// No matching invoice is a normal outcome (e.g. spontaneous deposits).
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil
		}
		return findErr
	}

	// Atomically mark the invoice as paid and link the payment to it.
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Invoice{}).
			Where("invoice_id = ?", invoice.InvoiceID).
			Where("invoice_status IN ?", []string{xrplPaymentStatusCreated, xrplPaymentStatusPending}).
			Update("invoice_status", xrplPaymentStatusPaid).Error; err != nil {
			return err
		}

		return tx.Model(&models.XRPLPayment{}).
			Where("tx_hash = ?", txHash).
			Update("invoice_id", invoice.InvoiceID).Error
	})
}

// fetchAccountTx calls the XRPL account_tx RPC method and returns the parsed
// result. It handles pagination by accepting an optional Marker from the
// previous call. ledgerIndexMin is inclusive; pass 0 to start from the
// beginning of ledger history.
func (r *XRPLReconciler) fetchAccountTx(address string, ledgerIndexMin int64, marker interface{}) (*xrplAccountTxResult, error) {
	params := map[string]interface{}{
		"account":          address,
		"ledger_index_min": ledgerIndexMin,
		"ledger_index_max": -1, // -1 means "up to the latest validated ledger"
		"binary":           false,
		"forward":          true, // oldest-first ordering keeps checkpoint updates monotonic
		"limit":            xrplAccountTxPageSizeLimit,
	}
	if marker != nil {
		params["marker"] = marker
	}

	requestBody, err := json.Marshal(xrplRPCRequest{
		Method: "account_tx",
		Params: []interface{}{params},
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, r.rpcURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := r.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("rpc status %d", response.StatusCode)
	}

	var rpcResponse xrplRPCResponse
	if err := json.NewDecoder(response.Body).Decode(&rpcResponse); err != nil {
		return nil, err
	}

	if rpcResponse.Result.Error != "" {
		return nil, fmt.Errorf("rpc error: %s", rpcResponse.Result.Error)
	}

	if rpcResponse.Result.Status != "" && rpcResponse.Result.Status != "success" {
		return nil, fmt.Errorf("rpc non-success status: %s", rpcResponse.Result.Status)
	}

	return &rpcResponse.Result, nil
}

// getCheckpoint returns the persisted ledger checkpoint for the given account.
// If no checkpoint exists yet (first run) it creates one with LastLedgerIndex=0
// so the reconciler scans from the beginning of history.
func (r *XRPLReconciler) getCheckpoint(account string) (*models.XRPLCheckpoint, error) {
	var checkpoint models.XRPLCheckpoint
	err := r.db.Where("account = ?", account).First(&checkpoint).Error
	if err == nil {
		return &checkpoint, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	checkpoint = models.XRPLCheckpoint{
		Account:         account,
		LastLedgerIndex: 0,
		UpdatedAt:       time.Now().UTC(),
	}

	if createErr := r.db.Create(&checkpoint).Error; createErr != nil {
		return nil, createErr
	}

	return &checkpoint, nil
}

// updateCheckpoint advances the stored last-seen ledger index for the given
// account. This is called at the end of each successful reconciliation pass so
// the next run only fetches transactions from new ledgers.
func (r *XRPLReconciler) updateCheckpoint(account string, ledgerIndex int64) error {
	return r.db.Model(&models.XRPLCheckpoint{}).
		Where("account = ?", account).
		Updates(map[string]interface{}{
			"last_ledger_index": ledgerIndex,
			"updated_at":        time.Now().UTC(),
		}).Error
}
