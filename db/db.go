package db

import (
	"database/sql"
	"errors"
	"time"
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

// Retrieve leaderboard, ordered by amount locked
func (db *DBParams) GetLeaderboard() ([]LeaderboardEntry, error) {
	rows, err := db.DB.Query(`SELECT id, amount_locked, paymail, public_key, created_at, updated_at FROM leaderboard ORDER BY amount_locked DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.ID, &entry.AmountLocked, &entry.Paymail, &entry.PublicKey, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// Update verified user with additional fields
func (db *DBParams) UpdateVerifiedUser(paymail, telegramUsername, challenge, signature string) error {
	_, err := db.DB.Exec(`UPDATE leaderboard SET telegram_username=$1, is_verified=$2, challenge=$3, signature=$4, updated_at=$5 WHERE paymail=$6`,
		telegramUsername, true, challenge, signature, time.Now(), paymail)
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

func (db *DBParams) UpsertUser(amountLocked float64, paymail, pubkey string) error {
	// Prepare SQL for upsert
	sqlStatement := `
			INSERT INTO leaderboard	(amount_locked, paymail, public_key, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (paymail)
			DO UPDATE SET amount_locked = EXCLUDED.amount_locked, updated_at = NOW();
	`

	_, err := db.DB.Exec(sqlStatement, amountLocked, paymail, pubkey, time.Now(), time.Now())
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
