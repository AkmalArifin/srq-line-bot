package routes

import (
	"example.com/yahfaz/models"
	"example.com/yahfaz/utils"
)

func createMemorizationPage(pageID int64, userID string) error {

	var userMemorization models.UserMemorization
	userMemorization.UserID.SetValid(userID)
	userMemorization.CardType.SetValid("page")
	userMemorization.PageID.SetValid(pageID)
	userMemorization.Level.SetValid(1)
	userMemorization.TimeReview.SetValue(utils.GetTimeReview(1))
	userMemorization.RepetitionCount.SetValid(1)
	userMemorization.CorrectCount.SetValid(1)
	err := userMemorization.Save()

	if err != nil {
		return err
	}

	return nil
}

func reviewMemorization(id int64, ease string) error {
	userMemorization, err := models.GetUserMemorizationByID(id)

	if err != nil {
		return err
	}

	nextLevel := utils.GetNextLevel(userMemorization.Level.ValueOrZero(), ease)
	nextTimeReview := utils.GetTimeReview(nextLevel)
	userMemorization.Level.SetValid(nextLevel)
	userMemorization.TimeReview.SetValue(nextTimeReview)
	userMemorization.RepetitionCount.Int64 += 1

	err = userMemorization.Update()

	if err != nil {
		return err
	}

	return nil
}
