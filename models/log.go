package models

import (
	"time"

	"example.com/yahfaz/db"
	"github.com/guregu/null/v5"
)

type Log struct {
	ID           int64       `json:"id"`
	CardType     null.String `json:"card_type"`
	LearningType null.String `json:"learning_type"`
	UserID       null.String `json:"user_id"`
	AyahID       null.Int    `json:"ayah_id"`
	PageID       null.Int    `json:"page_id"`
	Ease         null.Int    `json:"ease"`
	Level        null.Int    `json:"level"`
	CreatedAt    NullTime    `json:"created_at"`
}

func GetAllLog() ([]Log, error) {
	query := "SELECT * FROM logs"
	rows, err := db.DB.Query(query)

	if err != nil {
		return nil, err
	}

	var logs []Log

	for rows.Next() {
		var log Log
		err = rows.Scan(&log.ID, &log.CardType, &log.LearningType, &log.UserID, &log.AyahID, &log.PageID, &log.Ease, &log.Level, &log.CreatedAt)

		if err != nil {
			return nil, err
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func GetLogByUserID(userID int64) ([]Log, error) {
	query := "SELECT * FROM logs WHERE user_id = ?"
	rows, err := db.DB.Query(query, userID)

	if err != nil {
		return nil, err
	}

	var logs []Log

	for rows.Next() {
		var log Log
		err = rows.Scan(&log.ID, &log.CardType, &log.LearningType, &log.UserID, &log.AyahID, &log.PageID, &log.Ease, &log.Level, &log.CreatedAt)

		if err != nil {
			return nil, err
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func (l *Log) Save() error {
	query := `
	INSERT INTO log (card_type, learning_type, user_id, ayah_id, page_id, ease, level, time_review, repetition_count, correct_count, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	stmt, err := db.DB.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	l.CreatedAt.SetValue(time.Now())

	results, err := stmt.Exec(l.CardType, l.LearningType, l.UserID, l.AyahID, l.PageID, l.Ease, l.Level, l.CreatedAt)

	if err != nil {
		return err
	}

	l.ID, err = results.LastInsertId()

	return err
}
