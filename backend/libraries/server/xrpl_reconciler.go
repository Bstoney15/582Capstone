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
	defaultXRPLRPCURL          = "https://s.altnet.rippletest.net:51234"
	defaultReconcileInterval   = 5 * time.Second
	xrplPaymentStatusCreated   = "created"
	xrplPaymentStatusPending   = "verification_pending"
	xrplPaymentStatusPaid      = "paid"
	xrplAccountTxPageSizeLimit = 200
)

type xrplRPCRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type xrplRPCResponse struct {
	Result xrplAccountTxResult `json:"result"`
}

type xrplAccountTxResult struct {
	Transactions   []xrplTransactionEnvelope `json:"transactions"`
	LedgerIndexMax int64                     `json:"ledger_index_max"`
	Marker         interface{}               `json:"marker"`
	Status         string                    `json:"status"`
	Error          string                    `json:"error"`
}

type xrplTransactionEnvelope struct {
	Validated bool       `json:"validated"`
	Tx        xrplTx     `json:"tx"`
	TxJSON    xrplTx     `json:"tx_json"`
	Meta      xrplTxMeta `json:"meta"`
}

type xrplTxMeta struct {
	TransactionResult string `json:"TransactionResult"`
}

type xrplTx struct {
	TransactionType string      `json:"TransactionType"`
	Destination     string      `json:"Destination"`
	Amount          interface{} `json:"Amount"`
	Hash            string      `json:"hash"`
	LedgerIndex     int64       `json:"ledger_index"`
	DestinationTag  *uint32     `json:"DestinationTag"`
}

type XRPLReconciler struct {
	db            *gorm.DB
	httpClient    *http.Client
	rpcURL        string
	interval      time.Duration
	isReconciling bool
}

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

func (r *XRPLReconciler) Start() {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		r.runOnce()
		for range ticker.C {
			r.runOnce()
		}
	}()
}

func (r *XRPLReconciler) runOnce() {
	if r.isReconciling {
		return
	}
	r.isReconciling = true
	defer func() { r.isReconciling = false }()

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
			tx := envelope.Tx
			if tx.TransactionType == "" {
				tx = envelope.TxJSON
			}

			if !envelope.Validated {
				continue
			}

			if tx.TransactionType != "Payment" {
				continue
			}

			if envelope.Meta.TransactionResult != "tesSUCCESS" {
				continue
			}

			if tx.Destination != address {
				continue
			}

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

	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&payment)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return nil
	}

	dropsDecimal, err := decimal.NewFromString(amountDrops)
	if err != nil {
		return err
	}

	amountXRP := dropsDecimal.Div(decimal.NewFromInt(1_000_000)).Round(4)

	var invoice models.Invoice
	findErr := r.db.
		Joins("JOIN merchant_customers ON merchant_customers.customer_id = invoice.invoice_customer_id").
		Where("merchant_customers.customer_merchant_id = ?", merchantID).
		Where("invoice.invoice_status IN ?", []string{xrplPaymentStatusCreated, xrplPaymentStatusPending}).
		Where("invoice.invoice_amount_charged = ?", amountXRP).
		Order("invoice.invoice_id ASC").
		First(&invoice).Error

	if findErr != nil {
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil
		}
		return findErr
	}

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

func (r *XRPLReconciler) fetchAccountTx(address string, ledgerIndexMin int64, marker interface{}) (*xrplAccountTxResult, error) {
	params := map[string]interface{}{
		"account":          address,
		"ledger_index_min": ledgerIndexMin,
		"ledger_index_max": -1,
		"binary":           false,
		"forward":          true,
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

func (r *XRPLReconciler) updateCheckpoint(account string, ledgerIndex int64) error {
	return r.db.Model(&models.XRPLCheckpoint{}).
		Where("account = ?", account).
		Updates(map[string]interface{}{
			"last_ledger_index": ledgerIndex,
			"updated_at":        time.Now().UTC(),
		}).Error
}
