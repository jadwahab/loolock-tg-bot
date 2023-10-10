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

type DBParams struct {
	DB *sql.DB
}

type LeaderboardEntry struct {
	ID               int64
	AmountLocked     float64
	Paymail          string
	PublicKey        string
	TelegramUsername string
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
	return err
}

func (db *DBParams) RemoveUserFromGroupChatDB(chatID int64, userID int64) error {
	_, err := db.DB.Exec(
		"DELETE FROM group_chat_users WHERE chat_id = $1 AND user_id = $2",
		chatID, userID,
	)
	return err
}

// Retrieve top 100 leaderboard entries, ordered by amount locked
func (db *DBParams) GetLeaderboard() ([]LeaderboardEntry, error) {
	rows, err := db.DB.Query(`SELECT id, amount_locked, paymail, created_at, updated_at FROM leaderboard ORDER BY amount_locked DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.ID, &entry.AmountLocked, &entry.Paymail, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// Check if a specific paymail is in the top 100
func (db *DBParams) IsPaymailInTop100(paymail string) (bool, error) {
	// Construct a subquery to get the top 100 paymails
	subquery := `SELECT paymail FROM leaderboard ORDER BY amount_locked DESC LIMIT 100`

	// Use the subquery to directly check if the specific paymail is in the top 100
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM (%s) AS sub WHERE paymail = $1)`, subquery)

	var exists bool
	err := db.DB.QueryRow(query, paymail).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Update verified user with additional fields
func (db *DBParams) UpdateVerifiedUser(paymail, telegramUsername, challenge, pubkey, signature string) error {
	_, err := db.DB.Exec(`UPDATE leaderboard SET telegram_username=$1, is_verified=$2, challenge=$3, public_key=$4, signature=$5, updated_at=$6 WHERE paymail=$7`,
		telegramUsername, true, challenge, pubkey, signature, time.Now(), paymail)
	return err
}

// Delete leaderboard entry by ID
func (db *DBParams) DeleteEntryByID(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM leaderboard WHERE id=$1`, id)
	return err
}

// Find leaderboard entry by paymail
func (db *DBParams) FindEntryByPaymail(paymail string) (*LeaderboardEntry, error) {
	row := db.DB.QueryRow(`SELECT id, amount_locked, paymail, public_key, created_at, updated_at FROM leaderboard WHERE paymail=$1`, paymail)

	var entry LeaderboardEntry
	if err := row.Scan(&entry.ID, &entry.AmountLocked, &entry.Paymail, &entry.PublicKey, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
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
		VALUES %s
		ON CONFLICT (paymail)
		DO UPDATE SET amount_locked = EXCLUDED.amount_locked, updated_at = NOW();
	`
	sqlStatement = fmt.Sprintf(sqlStatement, strings.Join(valueStrings, ","))
	_, err := db.DB.Exec(sqlStatement, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (db *DBParams) GetUserByTelegramUsername(username string) (*LeaderboardEntry, error) {
	var user LeaderboardEntry
	if err := db.DB.QueryRow("SELECT * FROM leaderboard WHERE telegram_username = $1", username).Scan(&user.ID, &user.AmountLocked, &user.Paymail, &user.PublicKey, &user.TelegramUsername, &user.IsVerified, &user.Challenge, &user.Signature, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (db *DBParams) GetEntryByPaymail(paymail string) (*LeaderboardEntry, error) {
	var entry LeaderboardEntry
	if err := db.DB.QueryRow("SELECT * FROM leaderboard WHERE paymail = $1", paymail).Scan(&entry.ID, &entry.AmountLocked, &entry.Paymail, &entry.PublicKey, &entry.TelegramUsername, &entry.IsVerified, &entry.Challenge, &entry.Signature, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

func (db *DBParams) GetPaymailPubkey(paymail string) (string, error) {
	var pubkey string
	if err := db.DB.QueryRow("SELECT public_key FROM leaderboard WHERE paymail = $1", paymail).Scan(&pubkey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return pubkey, nil
}
