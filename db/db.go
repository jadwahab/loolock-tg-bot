package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jadwahab/loolock-tg-bot/apis"
	"github.com/tonicpow/go-paymail"
)

// TODO: separate db files for leaderboard and for chat db

type DBParams struct {
	DB *sql.DB
}

// TODO: handle sql null strings
type LeaderboardEntry struct {
	ID               int64
	AmountLocked     float64
	Paymail          string
	PublicKey        string
	TelegramUsername sql.NullString
	TelegramID       int64
	IsVerified       bool
	Challenge        string
	Signature        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (db *DBParams) AddUserToGroupChatDB(chatID int64, userID int64, username string) error {
	_, err := db.DB.Exec(
		"INSERT INTO group_chat_users (chat_id, user_id, username) VALUES ($1, $2, $3)",
		chatID, userID, username,
	)
	fmt.Printf("Added user %s, %d to chat %d", username, userID, chatID)
	return err
}

func (db *DBParams) RemoveUserFromGroupChatDB(chatID int64, userID int64) error {
	_, err := db.DB.Exec(
		"DELETE FROM group_chat_users WHERE chat_id = $1 AND user_id = $2",
		chatID, userID,
	)
	return err
}

// Check if user exists in table
func (db *DBParams) UserExists(chatID int64, userID int64) (bool, error) {
	var exists bool

	query := `SELECT exists(SELECT 1 FROM group_chat_users WHERE chat_id=$1 AND user_id=$2)`
	err := db.DB.QueryRow(query, chatID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

type ChatUser struct {
	UserID   int64
	UserName string
}

func (db *DBParams) GetUniqueUsers(chatID int64) ([]ChatUser, error) {
	var users []ChatUser

	rows, err := db.DB.Query("SELECT DISTINCT user_id, username FROM group_chat_users WHERE chat_id = $1", chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user ChatUser
		if err := rows.Scan(&user.UserID, &user.UserName); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Retrieve top leaderboard entries, ordered by amount locked
func (db *DBParams) GetLeaderboard(limit int) ([]LeaderboardEntry, error) {
	rows, err := db.DB.Query(fmt.Sprintf(`SELECT amount_locked, paymail, telegram_username, is_verified FROM leaderboard ORDER BY amount_locked DESC LIMIT %d`, limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.AmountLocked, &entry.Paymail, &entry.TelegramUsername, &entry.IsVerified); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// Update verified user with additional fields
func (db *DBParams) UpdateVerifiedUser(paymail, telegramUsername, challenge, pubkey, signature string, telegram_id int64) error {
	_, err := db.DB.Exec(`UPDATE leaderboard SET telegram_username=$1, telegram_id=$2, is_verified=$3, challenge=$4, public_key=$5, signature=$6, updated_at=$7 WHERE paymail=$8`,
		telegramUsername, telegram_id, true, challenge, pubkey, signature, time.Now(), paymail)
	return err
}

func (db *DBParams) UpsertUser(amountLocked float64, paymail string) error {
	// Prepare SQL for upsert
	sqlStatement := `
			INSERT INTO leaderboard	(amount_locked, paymail, created_at, updated_at) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (paymail)
			DO UPDATE SET amount_locked = EXCLUDED.amount_locked, updated_at = NOW();
	`

	_, err := db.DB.Exec(sqlStatement, amountLocked, paymail, time.Now(), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (db *DBParams) BatchUpsert(bitcoiners []apis.Bitcoiner) error {
	// Prepare data
	var valueStrings []string
	var valueArgs []interface{}
	i := 1
	for _, bitcoiner := range bitcoiners {
		s, err := paymail.ValidateAndSanitisePaymail("1"+bitcoiner.Handle, false) // TODO: harden instead of prepending '1'
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i, i+1, i+2, i+3))
		valueArgs = append(valueArgs, bitcoiner.TotalAmountLocked, s.Address, time.Now(), time.Now())
		i += 4
	}

	sqlStatement := `
		INSERT INTO leaderboard (amount_locked, paymail, created_at, updated_at) 
		VALUES ` + strings.Join(valueStrings, ",") + `
		ON CONFLICT (paymail)
		DO UPDATE SET amount_locked = EXCLUDED.amount_locked, updated_at = NOW();
	`
	_, err := db.DB.Exec(sqlStatement, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (db *DBParams) GetUserByTelegramUsername(username string) (*LeaderboardEntry, error) {
	var user LeaderboardEntry
	if err := db.DB.QueryRow("SELECT id, amount_locked, paymail, is_verified FROM leaderboard WHERE telegram_username = $1", username).Scan(&user.ID, &user.AmountLocked, &user.Paymail, &user.IsVerified); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (db *DBParams) GetUserByTelegramID(id int64) (*LeaderboardEntry, error) {
	var user LeaderboardEntry
	if err := db.DB.QueryRow("SELECT id, amount_locked, paymail, is_verified FROM leaderboard WHERE telegram_id = $1", id).Scan(&user.ID, &user.AmountLocked, &user.Paymail, &user.IsVerified); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (db *DBParams) PaymailExists(paymail string) (bool, error) {
	var exists bool

	query := `SELECT exists(SELECT 1 FROM leaderboard WHERE paymail=$1)`
	err := db.DB.QueryRow(query, paymail).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// func (db *DBParams) UpsertChallengeAndUsername(challengeValue, telegramUsernameValue string) error {
// 	sqlStatement := `
// 		INSERT INTO leaderboard (challenge, telegram_username)
// 		VALUES ($1, $2)
// 		ON CONFLICT (telegram_username)
// 		DO UPDATE SET challenge = EXCLUDED.challenge;
// 	`
// 	_, err := db.DB.Exec(sqlStatement, challengeValue, telegramUsernameValue)
// 	return err
// }

// func (db *DBParams) GetChallengeForPaymail(paymailValue string) (string, error) {
// 	var challenge string
// 	sqlStatement := `
// 		SELECT challenge
// 		FROM leaderboard
// 		WHERE paymail = $1;
// 	`
// 	err := db.DB.QueryRow(sqlStatement, paymailValue).Scan(&challenge)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return "", nil
// 		}
// 		return "", err
// 	}

// 	return challenge, nil
// }

// func (db *DBParams) GetEntryByPaymail(paymail string) (*LeaderboardEntry, error) {
// 	var entry LeaderboardEntry
// 	if err := db.DB.QueryRow("SELECT * FROM leaderboard WHERE paymail = $1", paymail).Scan(&entry.ID, &entry.AmountLocked, &entry.Paymail, &entry.PublicKey, &entry.TelegramUsername, &entry.IsVerified, &entry.Challenge, &entry.Signature, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &entry, nil
// }

// func (db *DBParams) GetPaymailPubkey(paymail string) (string, error) {
// 	var pubkey string
// 	if err := db.DB.QueryRow("SELECT public_key FROM leaderboard WHERE paymail = $1", paymail).Scan(&pubkey); err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return "", nil
// 		}
// 		return "", err
// 	}
// 	return pubkey, nil
// }

// // Delete leaderboard entry by ID
// func (db *DBParams) DeleteEntryByID(id int64) error {
// 	_, err := db.DB.Exec(`DELETE FROM leaderboard WHERE id=$1`, id)
// 	return err
// }

// // Find leaderboard entry by paymail
// func (db *DBParams) FindEntryByPaymail(paymail string) (*LeaderboardEntry, error) {
// 	row := db.DB.QueryRow(`SELECT id, amount_locked, paymail, public_key, created_at, updated_at FROM leaderboard WHERE paymail=$1`, paymail)

// 	var entry LeaderboardEntry
// 	if err := row.Scan(&entry.ID, &entry.AmountLocked, &entry.Paymail, &entry.PublicKey, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &entry, nil
// }

// // Check if a specific paymail is in the top 100
// func (db *DBParams) IsPaymailInTop100(paymail string) (bool, error) {
// 	// Construct a subquery to get the top 100 paymails
// 	subquery := `SELECT paymail FROM leaderboard ORDER BY amount_locked DESC LIMIT 100`

// 	// Use the subquery to directly check if the specific paymail is in the top 100
// 	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM (%s) AS sub WHERE paymail = $1)`, subquery)

// 	var exists bool
// 	err := db.DB.QueryRow(query, paymail).Scan(&exists)
// 	if err != nil {
// 		return false, err
// 	}
// 	return exists, nil
// }
