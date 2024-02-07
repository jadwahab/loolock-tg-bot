package db

import "log"

// FIXME: make the db primary key include the chatID if want to create multiple groups

func (db *DBParams) AddUserToGroupChatDB(chatID int64, userID int64, username, tgName string, member bool) error {
	_, err := db.DB.Exec(
		`INSERT INTO group_chat_users (chat_id, user_id, username, tg_name, member) 
		VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT (user_id) DO UPDATE 
		SET member = EXCLUDED.member`,
		chatID, userID, username, tgName, member,
	)
	if err != nil {
		log.Printf("Error adding/updating user %s, %d in chat %d: %v\n", username, userID, chatID, err)
	} else {
		log.Printf("Added/Updated user %s, %d in chat %d\n", username, userID, chatID)
	}
	return err
}

func (db *DBParams) UpdateUserLeftAt(chatID int64, userID int64) error {
	_, err := db.DB.Exec(
		"UPDATE group_chat_users SET left_at = CURRENT_TIMESTAMP, member = false WHERE chat_id = $1 AND user_id = $2",
		chatID, userID,
	)
	if err != nil {
		log.Printf("Error updating left_at and member for user %d in chat %d: %v\n", userID, chatID, err)
	} else {
		log.Printf("Updated left_at and member for user %d in chat %d\n", userID, chatID)
	}
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
