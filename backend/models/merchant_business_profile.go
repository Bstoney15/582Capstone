package models

// Author: Benjamin Stonestreet
// Created: 2024-02-02

// MerchantBusinessProfile stores the business details of a merchant, such as
// DBA name, tax ID, legal structure, and contact information.
type MerchantBusinessProfile struct {
	MerchantBusinessProfileID                 string   `gorm:"primaryKey;type:varchar(36)" json:"merchant_business_profile_id"`
	MerchantBusinessProfileMerchantID         string   `gorm:"not null;type:varchar(36)" json:"merchant_business_profile_merchant_id"`
	Merchant                                  Merchant `gorm:"foreignKey:MerchantBusinessProfileMerchantID;references:MerchantID" json:"merchant,omitempty"`
	MerchantBusinessProfileDBAName            string   `gorm:"not null" json:"merchant_business_profile_dba_name"`
	MerchantBusinessProfileRegistrationNumber string   `gorm:"not null" json:"merchant_business_profile_registration_number"`
	MerchantBusinessProfileTaxID              string   `gorm:"not null" json:"merchant_business_profile_tax_id"`
	MerchantBusinessProfileWebsiteURL         string   `gorm:"not null" json:"merchant_business_profile_website_url"`
	MerchantBusinessProfileIncoporationDate   string   `gorm:"not null" json:"merchant_business_profile_incoporation_date"`
	MerchantBusinessProfileLegalStructure     string   `gorm:"not null" json:"merchant_business_profile_legal_structure"`
	MerchantBusinessProfileMCCCode            string   `gorm:"not null" json:"merchant_business_profile_mcc_code"`
	MerchantBusinessProfilePhoneNumber        string   `gorm:"not null" json:"merchant_business_profile_phone_number"`
	MerchantBusinessProfileEmail              string   `gorm:"not null" json:"merchant_business_profile_email"`
}

// TableName overrides the default GORM table name to "merchant_business_profile".
func (MerchantBusinessProfile) TableName() string {
	return "merchant_business_profile"
}
