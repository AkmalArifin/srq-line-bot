package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

type NullTime struct {
	mysql.NullTime
}

func (nt *NullTime) SetValue(t time.Time) {
	nt.Time = t
	nt.Valid = true
}

func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time)
}

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == 'n' {
		nt.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &nt.Time); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	nt.Valid = true
	return nil
}
