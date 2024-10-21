package models

import (
	"time"

	"example.com/yahfaz/db"
	"github.com/guregu/null/v5"
)

type User struct {
	ID          int64       `json:"id"`
	UserID      null.String `json:"user_id"`
	DisplayName null.String `json:"display_name"`
	Language    null.String `json:"language"`
	CreatedAt   NullTime    `json:"created_at"`
}

func GetAllUser() ([]User, error) {
	query := "SELECT * FROM users"
	rows, err := db.DB.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.UserID, &user.DisplayName, &user.Language, &user.CreatedAt)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (u *User) Save() error {
	query := `
	INSERT INTO users (user_id, display_name, language, created_at)
	VALUES(?, ?, ?, ?)
	`

	stmt, err := db.DB.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	u.CreatedAt.SetValue(time.Now())

	results, err := stmt.Exec(u.UserID, u.DisplayName, u.Language, u.CreatedAt)

	if err != nil {
		return err
	}

	u.ID, err = results.LastInsertId()

	return err
}
