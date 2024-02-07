package helpers

import (
	"sort"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jadwahab/loolock-tg-bot/config"
	"github.com/jadwahab/loolock-tg-bot/db"
	"github.com/pkg/errors"
)

func HandleUserOverflow(dbp *db.DBParams, cfg config.Config, bot *tgbotapi.BotAPI, chatID int64, members []db.ChatUser, chatLimit int) error {
	if len(members) < chatLimit {
		return errors.New("members count is less than chat limit")
	}

	err := RefreshLeaderboard(dbp)
	if err != nil {
		return errors.Wrap(err, "failed to refresh leaderboard when adding user")
	}

	lbLimit := chatLimit * 5 // TODO: refactor to use config. this number is used to get verified leaderboard and compare with current members
	// last I checked on 07/02/24 there were only 63 verified members and some were so low I had to increase this limit or remove entirely
	lbes, err := dbp.GetLeaderboard(true, lbLimit, "both")
	if err != nil {
		return errors.Wrap(err, "failed to get leaderboard")
	}

	// Map ranks from lbes for quick lookup
	ranksMap := make(map[int64]int)
	for _, lbe := range lbes {
		ranksMap[lbe.TelegramID.Int64] = lbe.Rank
	}
	// Collect rankings for users in array1
	rankedUsers := make([]UserRank, 0)
	for _, member := range members {
		if rank, exists := ranksMap[member.UserID]; exists {
			rankedUsers = append(rankedUsers, UserRank{UserID: member.UserID, UserName: member.UserName, Rank: rank})
		}
	}

	skipTop := len(members) - chatLimit
	// Sort the rankedUsers slice based on Rank
	sort.Slice(rankedUsers, func(i, j int) bool {
		return rankedUsers[i].Rank < rankedUsers[j].Rank // Ascending order, lower rank number is higher rank
	})
	// Determine the range to capture based on skipTop
	startIndex := skipTop
	if startIndex > len(rankedUsers) {
		startIndex = len(rankedUsers) // Ensure startIndex does not exceed bounds
	}
	endIndex := startIndex + (len(members) - skipTop)
	if endIndex > len(rankedUsers) {
		endIndex = len(rankedUsers)
	}
	// Return the next highest-ranking members after skipping the top 'skipTop' members
	kickedMembers := rankedUsers[startIndex:endIndex]

	// Kick the overflowing members
	for _, km := range kickedMembers {
		if km.UserID != bot.Self.ID {
			KickUser(bot, &KickArgs{
				ChatID:       chatID,
				UserID:       km.UserID,
				UserName:     km.UserName,
				KickDuration: time.Duration(cfg.KickDuration),
				DBP:          dbp,
				KickMessage:  "You were kicked because the group has reached the maximum number of members!",
			})
		}
	}

	return nil
}

type UserRank struct {
	UserID   int64
	UserName string
	Rank     int
}
