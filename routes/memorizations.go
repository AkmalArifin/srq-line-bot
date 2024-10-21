package routes

import (
	"fmt"
	"math"
	"sort"
	"time"

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
	userMemorization.UniqueKey.SetValid("P" + fmt.Sprintf("%04d", pageID) + "_" + userID)
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

// TODO
func statusMemorization(userID string) (map[time.Time]map[int]int, error) {
	memorizations, err := models.GetStatusByUserID(userID)

	if err != nil {
		return nil, err
	}

	status := make(map[time.Time]map[int]int)
	// borderTime := time.Now().Local().AddDate(0, 0, 7).Truncate(24 * time.Hour)

	for i := 0; i < 7; i++ {
		timeKey := time.Now().Local().AddDate(0, 0, i).Truncate(24 * time.Hour)
		status[timeKey] = make(map[int]int)
	}

	for _, m := range memorizations {
		// if m.TimeReview.Time.After(borderTime) {
		// 	break
		// }

		timeKey := m.TimeReview.Time.Local().Truncate(24 * time.Hour)
		hourKey := m.TimeReview.Time.Local().Hour()

		_, ok := status[timeKey]
		if !ok {
			continue
		}

		_, ok = status[timeKey][hourKey]
		if !ok {
			status[timeKey][hourKey] = 0
		}

		status[timeKey][hourKey] += 1
	}

	return status, nil
}

func showMemorizationPage(userID string) (map[int][]int, error) {
	memorizations, err := models.GetUserMemorizationByUserID(userID)

	if err != nil {
		return nil, err
	}

	juzPages := make(map[int][]int)

	for _, value := range memorizations {
		if value.PageID.ValueOrZero() == 0 {
			continue
		}

		key := int(math.Ceil(float64(value.PageID.ValueOrZero()-1) / 20.0))

		if key < 1 {
			key = 1
		}
		if key > 30 {
			key = 30
		}

		_, ok := juzPages[key]
		if !ok {
			juzPages[key] = []int{}
		}
		juzPages[key] = append(juzPages[key], int(value.PageID.ValueOrZero()))
	}

	for k := range juzPages {
		sort.Ints(juzPages[k])
	}

	return juzPages, nil
}
