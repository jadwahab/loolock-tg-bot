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

// TODO: handle sql null strings
type LeaderboardEntry struct {
	ID               int64
	AmountLocked     float64
	Paymail          string
	PublicKey        string
	TelegramUsername sql.NullString
	TelegramID       sql.NullInt64
	IsVerified       bool
	Challenge        string
	Signature        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	AmountLiked      float64
}

func (db *DBParams) GetLeaderboard(valid bool, limit int, orderBy string) ([]LeaderboardEntry, error) {
	var rows *sql.Rows
	var err error

	// Construct the base query string
	baseQuery := `
	SELECT amount_locked, amount_liked, paymail, telegram_username, telegram_id, is_verified
	FROM leaderboard`

	// Add the WHERE clause if valid is true
	if valid {
		baseQuery += ` WHERE is_verified = true`
	}

	// Determine the ORDER BY clause based on the orderBy parameter
	switch orderBy {
	case "locked":
		baseQuery += ` ORDER BY amount_locked DESC`
	case "liked":
		baseQuery += ` ORDER BY amount_liked DESC`
	case "both":
		baseQuery += ` ORDER BY (amount_locked + amount_liked) DESC`
	default:
		return nil, errors.New("invalid orderBy parameter")
	}

	// Add the LIMIT clause if limit is greater than 0
	if limit > 0 {
		baseQuery += ` LIMIT $1`
		rows, err = db.DB.Query(baseQuery, limit)
	} else {
		rows, err = db.DB.Query(baseQuery)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.AmountLocked, &entry.AmountLiked, &entry.Paymail, &entry.TelegramUsername, &entry.TelegramID, &entry.IsVerified); err != nil {
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

func (db *DBParams) BatchUpsertLocked(bitcoiners []apis.LockedBitcoiner) error {
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

func (db *DBParams) BatchUpsertLiked(bitcoiners []apis.LikedBitcoiner) error {
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
		valueArgs = append(valueArgs, bitcoiner.TotalLockLikedFromOthers, s.Address, time.Now(), time.Now())
		i += 4
	}

	sqlStatement := `
		INSERT INTO leaderboard (amount_liked, paymail, created_at, updated_at) 
		VALUES ` + strings.Join(valueStrings, ",") + `
		ON CONFLICT (paymail)
		DO UPDATE SET amount_liked = EXCLUDED.amount_liked, updated_at = NOW();
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
