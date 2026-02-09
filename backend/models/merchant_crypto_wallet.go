package models

type MerchantCryptoWallet struct {
	MerchantCryptoWalletID         string   `gorm:"primaryKey" json:"merchant_crypto_wallet_id"`
	MerchantCryptoWalletMerchantID string   `gorm:"not null" json:"merchant_crypto_wallet_merchant_id"`
	Merchant                       Merchant `gorm:"foreignKey:MerchantCryptoWalletMerchantID;references:MerchantID" json:"merchant,omitempty"`
	MerchantCryptoWalletAddress    string   `gorm:"not null" json:"merchant_crypto_wallet_address"`
	MerchantCryptoWalletVerified   bool     `gorm:"not null" json:"merchant_crypto_wallet_verified"`
}

func (MerchantCryptoWallet) TableName() string {
	return "merchant_crypto_wallet"
}
