package db

import "fmt"

func (db *DBParams) AddUserToGroupChatDB(chatID int64, userID int64, username string) error {
	_, err := db.DB.Exec(
		"INSERT INTO group_chat_users (chat_id, user_id, username) VALUES ($1, $2, $3)",
		chatID, userID, username,
	)
	fmt.Printf("Added user %s, %d to chat %d\n", username, userID, chatID)
	return err
}

func (db *DBParams) UpdateUserLeftAt(chatID int64, userID int64) error {
	_, err := db.DB.Exec(
		"UPDATE group_chat_users SET left_at = CURRENT_TIMESTAMP WHERE chat_id = $1 AND user_id = $2",
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
