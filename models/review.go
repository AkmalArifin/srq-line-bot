package models

import "github.com/guregu/null/v5"

type Review struct {
	ID   null.Int64 `json:"id"`
	Ease null.Int   `json:"ease"`
}
