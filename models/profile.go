package models

import "github.com/guregu/null/v5"

type Profile struct {
	UserID        null.String `json:"userId"`
	DisplayName   null.String `json:"displayName"`
	StatusMessage null.String `json:"statusMessage"`
	Language      null.String `json:"language"`
}
