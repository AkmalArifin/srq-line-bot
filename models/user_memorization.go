package models

import (
	"time"

	"example.com/yahfaz/db"
	"github.com/guregu/null/v5"
)

type UserMemorization struct {
	ID              int64       `json:"id"`
	UserID          null.String `json:"user_id"`
	CardType        null.String `json:"card_type"`
	AyahID          null.Int64  `json:"ayah_id"`
	PageID          null.Int64  `json:"page_id"`
	Level           null.Int    `json:"level"`
	TimeReview      NullTime    `json:"time_review"`
	RepetitionCount null.Int    `json:"repetition_count"`
	CorrectCount    null.Int    `json:"correct_count"`
	CreatedAt       NullTime    `json:"created_at"`
	UpdatedAt       NullTime    `json:"updated_at"`
}

func GetAllUserMemorization() ([]UserMemorization, error) {
	query := "SELECT * FROM user_memorization"
	rows, err := db.DB.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var userMemorizations []UserMemorization

	for rows.Next() {
		var userMemorization UserMemorization
		err = rows.Scan(&userMemorization.ID, &userMemorization.UserID, &userMemorization.CardType, &userMemorization.AyahID, &userMemorization.PageID, &userMemorization.Level, &userMemorization.TimeReview, &userMemorization.RepetitionCount, &userMemorization.CorrectCount, &userMemorization.CreatedAt, &userMemorization.UpdatedAt)

		if err != nil {
			return nil, err
		}

		userMemorizations = append(userMemorizations, userMemorization)
	}

	return userMemorizations, nil
}

func GetUserMemorizationByUserID(userID string) ([]UserMemorization, error) {
	query := "SELECT * FROM user_memorization WHERE user_id = ?"
	rows, err := db.DB.Query(query, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var userMemorizations []UserMemorization

	for rows.Next() {
		var userMemorization UserMemorization
		err = rows.Scan(&userMemorization.ID, &userMemorization.UserID, &userMemorization.CardType, &userMemorization.AyahID, &userMemorization.PageID, &userMemorization.Level, &userMemorization.TimeReview, &userMemorization.RepetitionCount, &userMemorization.CorrectCount, &userMemorization.CreatedAt, &userMemorization.UpdatedAt)

		if err != nil {
			return nil, err
		}

		userMemorizations = append(userMemorizations, userMemorization)
	}

	return userMemorizations, nil
}

func GetReviewByUserID(userID string) ([]UserMemorization, error) {
	query := "SELECT * FROM user_memorization WHERE user_id = ? AND time_review < ?"
	rows, err := db.DB.Query(query, userID, time.Now())

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var userMemorizations []UserMemorization

	for rows.Next() {
		var userMemorization UserMemorization
		err = rows.Scan(&userMemorization.ID, &userMemorization.UserID, &userMemorization.CardType, &userMemorization.AyahID, &userMemorization.PageID, &userMemorization.Level, &userMemorization.TimeReview, &userMemorization.RepetitionCount, &userMemorization.CorrectCount, &userMemorization.CreatedAt, &userMemorization.UpdatedAt)

		if err != nil {
			return nil, err
		}

		userMemorizations = append(userMemorizations, userMemorization)
	}

	return userMemorizations, nil
}

func GetUserMemorizationByID(id int64) (UserMemorization, error) {
	query := "SELECT * FROM user_memorization WHERE id = ?"
	row := db.DB.QueryRow(query, id)

	var userMemorization UserMemorization
	err := row.Scan(&userMemorization.ID, &userMemorization.UserID, &userMemorization.CardType, &userMemorization.AyahID, &userMemorization.PageID, &userMemorization.Level, &userMemorization.TimeReview, &userMemorization.RepetitionCount, &userMemorization.CorrectCount, &userMemorization.CreatedAt, &userMemorization.UpdatedAt)

	if err != nil {
		return UserMemorization{}, err
	}

	return userMemorization, nil
}

func (m *UserMemorization) Save() error {
	query := `
	INSERT INTO user_memorization (user_id, card_type, ayah_id, page_id, level, time_review, repetition_count, correct_count, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	stmt, err := db.DB.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	m.CreatedAt.SetValue(time.Now())
	m.UpdatedAt.SetValue(time.Now())

	results, err := stmt.Exec(m.UserID, m.CardType, m.AyahID, m.PageID, m.Level, m.TimeReview, m.RepetitionCount, m.CorrectCount, m.CreatedAt, m.UpdatedAt)

	if err != nil {
		return err
	}

	m.ID, err = results.LastInsertId()

	return err
}

func (m *UserMemorization) Update() error {
	query := `
	UPDATE user_memorization
	SET level = ?, time_review = ?, repetition_count = ?, correct_count = ?, updated_at = ?
	WHERE id = ?
	`
	stmt, err := db.DB.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	m.UpdatedAt.SetValue(time.Now())

	_, err = stmt.Exec(m.Level, m.TimeReview, m.RepetitionCount, m.CorrectCount, m.UpdatedAt, m.ID)

	return err
}
