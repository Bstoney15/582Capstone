package server

import (
	"backend/models"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	seedMerchantID        = "seed-merchant-001"
	seedMerchantID2       = "seed-merchant-002"
	seedAdminUserID       = "seed-user-admin-001"
	seedDeveloperUserID   = "seed-user-dev-001"
	seedOwnerUserID       = "seed-user-owner-001"
	seedAdminRoleID       = "seed-role-admin-001"
	seedDeveloperRoleID   = "seed-role-dev-001"
	seedOwnerRoleID       = "seed-role-owner-001"
	seedAdminRoleID2      = "seed-role-admin-002"
	seedDeveloperRoleID2  = "seed-role-dev-002"
	seedOwnerRoleID2      = "seed-role-owner-002"
	seedCustomerID        = "seed-customer-001"
	seedDepositID         = "seed-deposit-001"
	seedInvoiceID         = "seed-invoice-001"
	seedMerchantAPIKeyID  = "seed-merchant-api-key-001"
	seedMerchantAPIKeyID2 = "seed-merchant-api-key-002"
	seedMerchantAddressID = "seed-merchant-address-001"
	seedMerchantAddressID2 = "seed-merchant-address-002"
	seedMerchantProfileID = "seed-merchant-profile-001"
	seedMerchantProfileID2 = "seed-merchant-profile-002"
	seedMerchantOwnerID   = "seed-merchant-owner-001"
	seedMerchantOwnerID2  = "seed-merchant-owner-002"
	seedMerchantWalletID  = "seed-merchant-wallet-001"
	seedMerchantWalletID2 = "seed-merchant-wallet-002"
	seedMerchantWebhookID = "seed-merchant-webhook-001"
	seedMerchantWebhookID2 = "seed-merchant-webhook-002"
	seedAdminUsername     = "dev_admin"
	seedDeveloperUsername = "dev_developer"
	seedOwnerUsername     = "dev_owner"
	seedUserPasswordPlain = "password"
	seedApiKeyHash        = "dev_api_key_hash"
	seedWebhookKey        = "dev_webhook_key"
	seedWalletAddress     = "rG3BgRh1xGaySMtvLoz35UZdTg15Htgi6j"
	seedDateLayout        = "2006-01-02"
)

func (s *Server) ResetSeedData() error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("deposit_id = ?", seedDepositID).Delete(&models.Deposit{}).Error; err != nil {
			return err
		}

		if err := tx.Where("invoice_id = ?", seedInvoiceID).Delete(&models.Invoice{}).Error; err != nil {
			return err
		}

		if err := tx.Where("customer_id = ?", seedCustomerID).Delete(&models.Customer{}).Error; err != nil {
			return err
		}

		if err := tx.Where("role_id IN ?", []string{seedAdminRoleID, seedDeveloperRoleID, seedOwnerRoleID, seedAdminRoleID2, seedDeveloperRoleID2, seedOwnerRoleID2}).Delete(&models.Role{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id IN ?", []string{seedAdminUserID, seedDeveloperUserID, seedOwnerUserID}).Delete(&models.User{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_webhook_key_id = ?", seedMerchantWebhookID).Delete(&models.MerchantWebhookKey{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_webhook_key_id = ?", seedMerchantWebhookID2).Delete(&models.MerchantWebhookKey{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_crypto_wallet_id = ?", seedMerchantWalletID).Delete(&models.MerchantCryptoWallet{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_crypto_wallet_id = ?", seedMerchantWalletID2).Delete(&models.MerchantCryptoWallet{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_owner_id = ?", seedMerchantOwnerID).Delete(&models.MerchantOwner{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_owner_id = ?", seedMerchantOwnerID2).Delete(&models.MerchantOwner{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_business_profile_id = ?", seedMerchantProfileID).Delete(&models.MerchantBusinessProfile{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_business_profile_id = ?", seedMerchantProfileID2).Delete(&models.MerchantBusinessProfile{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_address_id = ?", seedMerchantAddressID).Delete(&models.MerchantAddress{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_address_id = ?", seedMerchantAddressID2).Delete(&models.MerchantAddress{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_api_key_id = ?", seedMerchantAPIKeyID).Delete(&models.MerchantAPIKey{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_api_key_id = ?", seedMerchantAPIKeyID2).Delete(&models.MerchantAPIKey{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_id = ?", seedMerchantID).Delete(&models.Merchant{}).Error; err != nil {
			return err
		}

		if err := tx.Where("merchant_id = ?", seedMerchantID2).Delete(&models.Merchant{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *Server) SeedDevData() error {
	adminPasswordHash, err := bcrypt.GenerateFromPassword([]byte(seedUserPasswordPlain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	developerPasswordHash, err := bcrypt.GenerateFromPassword([]byte(seedUserPasswordPlain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	ownerPasswordHash, err := bcrypt.GenerateFromPassword([]byte(seedUserPasswordPlain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	ownerDOB, err := time.Parse(seedDateLayout, "1990-01-01")
	if err != nil {
		return err
	}

	ownerStake, err := decimal.NewFromString("100.0000")
	if err != nil {
		return err
	}

	depositAmount, err := decimal.NewFromString("245.7500")
	if err != nil {
		return err
	}

	depositFeeAmount, err := decimal.NewFromString("2.5000")
	if err != nil {
		return err
	}

	invoiceAmount, err := decimal.NewFromString("120.0000")
	if err != nil {
		return err
	}

	invoiceFeeAmount, err := decimal.NewFromString("1.2000")
	if err != nil {
		return err
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		merchant := models.Merchant{
			MerchantID:   seedMerchantID,
			MerchantName: "Coffee Shop POS",
		}
		if err := upsert(tx, &merchant); err != nil {
			return err
		}

		merchant2 := models.Merchant{
			MerchantID:   seedMerchantID2,
			MerchantName: "Acai Bowl POS",
		}
		if err := upsert(tx, &merchant2); err != nil {
			return err
		}

		merchantAPIKey := models.MerchantAPIKey{
			MerchantAPIKeyID:         seedMerchantAPIKeyID,
			MerchantAPIKeyHashed:     seedApiKeyHash,
			MerchantAPIKeyMerchantID: seedMerchantID,
		}
		if err := upsert(tx, &merchantAPIKey); err != nil {
			return err
		}

		merchantAPIKey2 := models.MerchantAPIKey{
			MerchantAPIKeyID:         seedMerchantAPIKeyID2,
			MerchantAPIKeyHashed:     seedApiKeyHash,
			MerchantAPIKeyMerchantID: seedMerchantID2,
		}
		if err := upsert(tx, &merchantAPIKey2); err != nil {
			return err
		}

		merchantAddress := models.MerchantAddress{
			MerchantAddressID:         seedMerchantAddressID,
			MerchantAddressMerchantID: seedMerchantID,
			MerchantAddressLine1:      "1 Developer Way",
			MerchantAddressLine2:      "Suite 100",
			MerchantAddressCity:       "Ann Arbor",
			MerchantAddressState:      "MI",
			MerchantAddressPostalCode: 48104,
			MerchantAddressVerified:   true,
		}
		if err := upsert(tx, &merchantAddress); err != nil {
			return err
		}

		merchantAddress2 := models.MerchantAddress{
			MerchantAddressID:         seedMerchantAddressID2,
			MerchantAddressMerchantID: seedMerchantID2,
			MerchantAddressLine1:      "2 Developer Way",
			MerchantAddressLine2:      "Suite 200",
			MerchantAddressCity:       "Ann Arbor",
			MerchantAddressState:      "MI",
			MerchantAddressPostalCode: 48104,
			MerchantAddressVerified:   true,
		}
		if err := upsert(tx, &merchantAddress2); err != nil {
			return err
		}

		merchantProfile := models.MerchantBusinessProfile{
			MerchantBusinessProfileID:                 seedMerchantProfileID,
			MerchantBusinessProfileMerchantID:         seedMerchantID,
			MerchantBusinessProfileDBAName:            "Seed Merchant",
			MerchantBusinessProfileRegistrationNumber: "REG-001",
			MerchantBusinessProfileTaxID:              "TAX-001",
			MerchantBusinessProfileWebsiteURL:         "https://seed-merchant.test",
			MerchantBusinessProfileIncoporationDate:   "2020-01-01",
			MerchantBusinessProfileLegalStructure:     "LLC",
			MerchantBusinessProfileMCCCode:            "5734",
			MerchantBusinessProfilePhoneNumber:        "+1-734-555-0100",
			MerchantBusinessProfileEmail:              "ops@seed-merchant.test",
		}
		if err := upsert(tx, &merchantProfile); err != nil {
			return err
		}

		merchantProfile2 := models.MerchantBusinessProfile{
			MerchantBusinessProfileID:                 seedMerchantProfileID2,
			MerchantBusinessProfileMerchantID:         seedMerchantID2,
			MerchantBusinessProfileDBAName:            "Seed Merchant Role Demo",
			MerchantBusinessProfileRegistrationNumber: "REG-002",
			MerchantBusinessProfileTaxID:              "TAX-002",
			MerchantBusinessProfileWebsiteURL:         "https://seed-merchant-role-demo.test",
			MerchantBusinessProfileIncoporationDate:   "2021-02-02",
			MerchantBusinessProfileLegalStructure:     "LLC",
			MerchantBusinessProfileMCCCode:            "5734",
			MerchantBusinessProfilePhoneNumber:        "+1-734-555-0200",
			MerchantBusinessProfileEmail:              "ops+demo@seed-merchant.test",
		}
		if err := upsert(tx, &merchantProfile2); err != nil {
			return err
		}

		merchantOwner := models.MerchantOwner{
			MerchantOwnerID:          seedMerchantOwnerID,
			MerchantOwnerMerchantID:  seedMerchantID,
			MerchantOwnerFirstName:   "Avery",
			MerchantOwnerLastName:    "Developer",
			MerchantOwnerPhoneNumber: "+1-734-555-0199",
			MerchantOwnerEmail:       "owner@seed-merchant.test",
			MerchantOwnerDOB:         ownerDOB,
			MerchantOwnerStake:       ownerStake,
		}
		if err := upsert(tx, &merchantOwner); err != nil {
			return err
		}

		merchantOwner2 := models.MerchantOwner{
			MerchantOwnerID:          seedMerchantOwnerID2,
			MerchantOwnerMerchantID:  seedMerchantID2,
			MerchantOwnerFirstName:   "Dev",
			MerchantOwnerLastName:    "Engineer",
			MerchantOwnerPhoneNumber: "+1-734-555-0210",
			MerchantOwnerEmail:       "developer-owner@seed-merchant.test",
			MerchantOwnerDOB:         ownerDOB,
			MerchantOwnerStake:       ownerStake,
		}
		if err := upsert(tx, &merchantOwner2); err != nil {
			return err
		}

		merchantWallet := models.MerchantCryptoWallet{
			MerchantCryptoWalletID:         seedMerchantWalletID,
			MerchantCryptoWalletMerchantID: seedMerchantID,
			MerchantCryptoWalletAddress:    seedWalletAddress,
			MerchantCryptoWalletVerified:   true,
		}
		if err := upsert(tx, &merchantWallet); err != nil {
			return err
		}

		merchantWallet2 := models.MerchantCryptoWallet{
			MerchantCryptoWalletID:         seedMerchantWalletID2,
			MerchantCryptoWalletMerchantID: seedMerchantID2,
			MerchantCryptoWalletAddress:    seedWalletAddress,
			MerchantCryptoWalletVerified:   true,
		}
		if err := upsert(tx, &merchantWallet2); err != nil {
			return err
		}

		merchantWebhook := models.MerchantWebhookKey{
			MerchantWebhookKeyID:         seedMerchantWebhookID,
			MerchantWebhookKey:           seedWebhookKey,
			MerchantWebhookKeyMerchantID: seedMerchantID,
		}
		if err := upsert(tx, &merchantWebhook); err != nil {
			return err
		}

		merchantWebhook2 := models.MerchantWebhookKey{
			MerchantWebhookKeyID:         seedMerchantWebhookID2,
			MerchantWebhookKey:           seedWebhookKey,
			MerchantWebhookKeyMerchantID: seedMerchantID2,
		}
		if err := upsert(tx, &merchantWebhook2); err != nil {
			return err
		}

		adminUser := models.User{
			UserID:           seedAdminUserID,
			UserUsername:     seedAdminUsername,
			UserFirstName:    "Dev",
			UserLastName:     "Admin",
			UserPasswordHash: string(adminPasswordHash),
			UserStatus:       "active",
		}
		if err := upsert(tx, &adminUser); err != nil {
			return err
		}

		developerUser := models.User{
			UserID:           seedDeveloperUserID,
			UserUsername:     seedDeveloperUsername,
			UserFirstName:    "Dev",
			UserLastName:     "Engineer",
			UserPasswordHash: string(developerPasswordHash),
			UserStatus:       "active",
		}
		if err := upsert(tx, &developerUser); err != nil {
			return err
		}

		ownerUser := models.User{
			UserID:           seedOwnerUserID,
			UserUsername:     seedOwnerUsername,
			UserFirstName:    "Morgan",
			UserLastName:     "Owner",
			UserPasswordHash: string(ownerPasswordHash),
			UserStatus:       "active",
		}
		if err := upsert(tx, &ownerUser); err != nil {
			return err
		}

		adminRole := models.Role{
			RoleID:         seedAdminRoleID,
			RoleMerchantID: seedMerchantID,
			RoleUserID:     seedAdminUserID,
			RoleName:       models.RoleAdmin,
		}
		if err := upsert(tx, &adminRole); err != nil {
			return err
		}

		developerRole := models.Role{
			RoleID:         seedDeveloperRoleID,
			RoleMerchantID: seedMerchantID,
			RoleUserID:     seedDeveloperUserID,
			RoleName:       models.RoleDeveloper,
		}
		if err := upsert(tx, &developerRole); err != nil {
			return err
		}

		ownerRole := models.Role{
			RoleID:         seedOwnerRoleID,
			RoleMerchantID: seedMerchantID,
			RoleUserID:     seedOwnerUserID,
			RoleName:       models.RoleOwner,
		}
		if err := upsert(tx, &ownerRole); err != nil {
			return err
		}

		adminRole2 := models.Role{
			RoleID:         seedAdminRoleID2,
			RoleMerchantID: seedMerchantID2,
			RoleUserID:     seedAdminUserID,
			RoleName:       models.RoleAdmin,
		}
		if err := upsert(tx, &adminRole2); err != nil {
			return err
		}

		developerRole2 := models.Role{
			RoleID:         seedDeveloperRoleID2,
			RoleMerchantID: seedMerchantID2,
			RoleUserID:     seedOwnerUserID,
			RoleName:       models.RoleDeveloper,
		}
		if err := upsert(tx, &developerRole2); err != nil {
			return err
		}

		ownerRole2 := models.Role{
			RoleID:         seedOwnerRoleID2,
			RoleMerchantID: seedMerchantID2,
			RoleUserID:     seedDeveloperUserID,
			RoleName:       models.RoleOwner,
		}
		if err := upsert(tx, &ownerRole2); err != nil {
			return err
		}

		customer := models.Customer{
			CustomerID:         seedCustomerID,
			CustomerMerchantID: seedMerchantID,
			CustomerFirstName:  "Taylor",
			CustomerLastName:   "Customer",
			CustomerEmail:      "taylor.customer@testmail.dev",
		}
		if err := upsert(tx, &customer); err != nil {
			return err
		}

		deposit := models.Deposit{
			DepositID:         seedDepositID,
			DepositAmount:     depositAmount,
			DepositStatus:     "pending",
			DepositFeeAmount:  depositFeeAmount,
			DepositFeeStatus:  "unpaid",
			DepositCryptoType: "USDC",
			DepositCustomerID: seedCustomerID,
		}
		if err := upsert(tx, &deposit); err != nil {
			return err
		}

		invoice := models.Invoice{
			InvoiceID:            seedInvoiceID,
			InvoiceAmountCharged: invoiceAmount,
			InvoiceStatus:        "created",
			InvoiceFeeAmount:     invoiceFeeAmount,
			InvoiceFeeStatus:     "unpaid",
			InvoiceCryptoType:    "USDC",
			InvoiceCustomerID:    seedCustomerID,
		}
		if err := upsert(tx, &invoice); err != nil {
			return err
		}

		log.Printf("Seeded merchants %s and %s with shared users and merchant-specific roles", seedMerchantID, seedMerchantID2)

		return nil
	})
}

func upsert(tx *gorm.DB, value interface{}) error {
	return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(value).Error
}
