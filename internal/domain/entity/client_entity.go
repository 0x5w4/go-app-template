package entity

import "time"

type Client struct {
	Base
	CompanyID             uint
	Company               *Company
	Code                  string
	Name                  string
	Phone                 string
	Fax                   *string
	Icon                  *string
	IconUpdatedAt         *time.Time
	DistrictID            uint
	District              *District
	Village               string
	PostalCode            string
	Address               string
	PICName               string
	PICPhone              string
	ClientSupportFeatures []*ClientSupportFeature
}

type ClientSupportFeature struct {
	ClientID         uint
	Client           *Client
	SupportFeatureID uint
	SupportFeature   *SupportFeature
	Order            int
}
