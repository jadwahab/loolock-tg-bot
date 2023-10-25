package helpers

import (
	"errors"

	"github.com/jadwahab/loolock-tg-bot/apis"
	"github.com/jadwahab/loolock-tg-bot/db"
)

func RefreshLeaderboard(dbp *db.DBParams) error {
	bitcoiners, err := apis.GetBitcoiners()
	if err != nil {
		return err
	}

	if len(bitcoiners) == 0 {
		return errors.New("error getting enough bitcoiners from API")
	}

	return dbp.BatchUpsert(bitcoiners)
}

func UserExistsInLeaderboard(leaderboard []db.LeaderboardEntry, userID int64) bool {
	for _, entry := range leaderboard {
		if entry.TelegramID == userID {
			return true
		}
	}
	return false
}
