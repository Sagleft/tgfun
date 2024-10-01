package tgfun

import (
	"database/sql"
	"errors"
	"strings"

	tb "gopkg.in/telebot.v3"
)

func isSQLErrNoRows(err error) bool {
	return err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows in result set")
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
