package helpers

import (
	"github.com/jadwahab/loolock-tg-bot/apis"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/pkg/errors"
)

func RefreshLeaderboard(dbp *db.DBParams) error {
	bitcoinersLocked, err := apis.GetBitcoinersLocked()

	if err != nil {
		return err
	}
	if len(bitcoinersLocked) == 0 {
		return errors.New("error getting enough bitcoiners locked from API")
	}
	err = dbp.BatchUpsertLocked(bitcoinersLocked)
	if err != nil {
		return errors.Wrap(err, "error upserting bitcoiners locked into DB")
	}

	bitcoinersLiked, err := apis.GetBitcoinersLiked()
	if err != nil {
		return err
	}
	if len(bitcoinersLocked) == 0 {
		return errors.New("error getting enough bitcoiners liked from API")
	}
	err = dbp.BatchUpsertLiked(bitcoinersLiked)
	if err != nil {
		return errors.Wrap(err, "error upserting bitcoiners liked into DB")
	}
	return nil
}

func UserExistsInLeaderboard(leaderboard []db.LeaderboardEntry, userID int64) bool {
	for _, entry := range leaderboard {
		if entry.TelegramID.Valid && entry.TelegramID.Int64 == userID {
			return true
		}
	}
	return false
}
