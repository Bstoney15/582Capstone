// merchant_crypto_wallet.go – GORM model storing a merchant's XRPL wallet address and verification status.
package models

// Author: Benjamin Stonestreet
// Created: 2026-02-09

// MerchantCryptoWallet holds the XRPL wallet address and verification status
// for a merchant to receive payments.
type MerchantCryptoWallet struct {
	MerchantCryptoWalletID         string   `gorm:"primaryKey;type:varchar(36)" json:"merchant_crypto_wallet_id"`
	MerchantCryptoWalletMerchantID string   `gorm:"not null;type:varchar(36)" json:"merchant_crypto_wallet_merchant_id"`
	Merchant                       Merchant `gorm:"foreignKey:MerchantCryptoWalletMerchantID;references:MerchantID" json:"merchant,omitempty"`
	MerchantCryptoWalletAddress    string   `gorm:"not null;uniqueIndex:ux_merchant_crypto_wallet_address" json:"merchant_crypto_wallet_address"`
	MerchantCryptoWalletVerified   bool     `gorm:"not null" json:"merchant_crypto_wallet_verified"`
}

// TableName sets the table name for MerchantCryptoWallet.
func (MerchantCryptoWallet) TableName() string {
	return "merchant_crypto_wallet"
}
