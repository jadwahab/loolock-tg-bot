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
	AmountLocked     int64
	Paymail          string
	PublicKey        string
	TelegramUsername string
	IsVerified       bool
	Challenge        string
	Signature        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Add a new leaderboard entry
func (db *DBParams) AddEntry(entry LeaderboardEntry) error {
	_, err := db.DB.Exec(`INSERT INTO leaderboard (amount_locked, paymail, public_key, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		entry.AmountLocked, entry.Paymail, entry.PublicKey, time.Now(), time.Now())
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

// Update leaderboard entry by ID
func (db *DBParams) UpdateEntryByID(id int64, updatedEntry LeaderboardEntry) error {
	_, err := db.DB.Exec(`UPDATE leaderboard SET amount_locked=$1, paymail=$2, public_key=$3, updated_at=$4 WHERE id=$5`,
		updatedEntry.AmountLocked, updatedEntry.Paymail, updatedEntry.PublicKey, time.Now(), id)
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

func (db *DBParams) UpdateLeaderboard(leaderboardData []LeaderboardEntry) error {
	// Prepare SQL for upsert
	sqlStatement := `
			INSERT INTO leaderboard (paymail, public_key, amount_locked, last_updated)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (paymail)
			DO UPDATE SET amount_locked = EXCLUDED.amount_locked, last_updated = NOW();
	`

	for _, entry := range leaderboardData {
		_, err := db.DB.Exec(sqlStatement, entry.Paymail, entry.PublicKey, entry.AmountLocked)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DBParams) GetUserByTelegramUsername(username string) (LeaderboardEntry, error) {
	var user LeaderboardEntry
	err := db.DB.QueryRow("SELECT * FROM leaderboard WHERE telegram_username = $1", username).Scan(&user.ID, &user.AmountLocked, &user.Paymail, &user.PublicKey, &user.TelegramUsername, &user.IsVerified, &user.Challenge, &user.Signature, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return LeaderboardEntry{}, err
	}
	return user, nil
}

func (db *DBParams) GetEntryByPaymail(paymail string) (LeaderboardEntry, error) {
	var entry LeaderboardEntry
	err := db.DB.QueryRow("SELECT * FROM leaderboard WHERE paymail = $1", paymail).Scan(&entry.ID, &entry.AmountLocked, &entry.Paymail, &entry.PublicKey, &entry.TelegramUsername, &entry.IsVerified, &entry.Challenge, &entry.Signature, &entry.CreatedAt, &entry.UpdatedAt)
	if err != nil {
		return LeaderboardEntry{}, err
	}
	return entry, nil
}
