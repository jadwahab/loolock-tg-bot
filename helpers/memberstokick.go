package helpers

import (
	"sort"

	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/pkg/errors"
)

type UserRank struct {
	UserID   int64
	UserName string
	Rank     int
}

func DetermineMembersToKick(dbp *db.DBParams, members []db.ChatUser, chatLimit int) ([]UserRank, error) {
	// err := RefreshLeaderboard(dbp)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to refresh leaderboard when handling user overflow")
	// }

	lbLimit := 400 // TODO: refactor to use config. this number is used to get verified leaderboard and compare with current members
	// last I checked on 07/02/24 there were only 63 verified members and some were so low I had to increase this limit or remove entirely
	lbes, err := dbp.GetLeaderboard(true, lbLimit, "both")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get leaderboard")
	}

	// Map ranks from lbes for quick lookup
	ranksMap := make(map[int64]int)
	for _, lbe := range lbes {
		ranksMap[lbe.TelegramID.Int64] = lbe.Rank
	}

	// Collect rankings for users in members array
	rankedUsers := make([]UserRank, 0)
	intruders := make([]UserRank, 0)
	for _, member := range members {
		if member.UserID == 6476402231 { // loolockbot
			continue
		}
		if rank, exists := ranksMap[member.UserID]; exists {
			rankedUsers = append(rankedUsers, UserRank{UserID: member.UserID, UserName: member.UserName, Rank: rank})
		} else {
			intruders = append(intruders, UserRank{UserID: member.UserID, UserName: member.UserName, Rank: 0})
		}
	}
	// Sort the rankedUsers slice based on Rank
	sort.Slice(rankedUsers, func(i, j int) bool {
		return rankedUsers[i].Rank < rankedUsers[j].Rank // Ascending order
	})

	// If the number of members is less than or equal to chatLimit, no one needs to be kicked
	if len(rankedUsers) <= chatLimit {
		return []UserRank{}, nil
	}

	// Return all users outside the first 'chatLimit' number of highest-ranked users
	membersToKick := append(rankedUsers[chatLimit:], intruders...)
	return membersToKick, nil
}
