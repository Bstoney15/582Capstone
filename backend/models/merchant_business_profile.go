package models

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

func (MerchantBusinessProfile) TableName() string {
	return "merchant_business_profile"
}
