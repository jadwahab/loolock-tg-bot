package db

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
