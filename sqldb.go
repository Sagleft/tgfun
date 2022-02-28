package tgfun

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	tb "github.com/Sagleft/telegobot"
)

func isSQLErrNoRows(err error) bool {
	return err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows in result set")
}

func closeRows(rows *sql.Rows) {
	if rows == nil {
		return
	}
	rows.Close()
}

func (uft *UsersFeature) getUserData(sender *tb.User) (*userData, error) {
	user, err := uft.getUserDBData(sender.ID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	u := userData{
		TelegramID: sender.ID,
		Name:       sender.FirstName + " " + sender.LastName,
		TgName:     "@" + sender.Username,
	}
	if u.Name == " " {
		u.Name = "anonymous"
	}
	err = uft.saveUser(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// returns nil when user not found
func (uft *UsersFeature) getUserDBData(telegramUsedID int64) (*userData, error) {
	user := &userData{
		TelegramID: telegramUsedID,
	}
	sqlQuery := "SELECT id,name FROM " + uft.TableName + " WHERE tid=? LIMIT 1"
	err := uft.DBConn.QueryRow(sqlQuery, telegramUsedID).Scan(
		&user.ID,
		&user.Name,
	)
	if err != nil {
		if isSQLErrNoRows(err) {
			return nil, nil
		}
		return nil, errors.New("failed to select user data: " + err.Error())
	}

	return user, nil
}

func (uft *UsersFeature) saveUser(user *userData) error {
	sqlQuery := "INSERT INTO " + uft.TableName + " SET tid=?, name=?, tgname=?"
	result, err := uft.DBConn.Exec(sqlQuery, user.TelegramID, user.Name, user.TgName)
	if err != nil {
		return errors.New("failed to save user: " + err.Error())
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to get rows affected count: " + err.Error())
	}
	if rowsAffected == 0 {
		return errors.New("failed to save user to db, 0 rows affected")
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return errors.New("failed to get saved user ID: " + err.Error())
	}
	user.ID = userID
	return nil
}

func (uft *UsersFeature) getUsersTelegramIDs() ([]int64, error) {
	sqlQuery := "SELECT tid FROM " + uft.TableName
	rows, err := uft.DBConn.Query(sqlQuery)
	defer closeRows(rows)
	if err != nil {
		return nil, errors.New("faile to select user IDs: " + err.Error())
	}

	ids := []int64{}
	var newID int64
	for rows.Next() {
		err := rows.Scan(&newID)
		if err != nil {
			log.Println("failed to scan user row: " + err.Error())
			continue
		}
		ids = append(ids, newID)
	}
	return ids, nil
}
