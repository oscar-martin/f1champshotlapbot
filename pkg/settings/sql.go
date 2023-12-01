package settings

import (
	"database/sql"
	"fmt"
)

func buildCreateNotificationsTable() string {
	return `CREATE TABLE IF NOT EXISTS notifications (
		userid TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		chatid TEXT NOT NULL,
		testday INTEGER,
		practice INTEGER,
		qual INTEGER,
		warnup INTEGER,
		race INTEGER);`
}

func buildSelectUserCommand(userID string) (string, func(*sql.Rows) (Notifications, error)) {
	fields := "testday, practice, qual, warnup, race"
	return fmt.Sprintf(`SELECT %s FROM notifications WHERE userid = '%s'`, fields, userID), processSelectUserRows
}

func processSelectUserRows(rows *sql.Rows) (Notifications, error) {
	defer rows.Close()

	n := AllDisabled()
	// only can be one row
	if rows.Next() {
		var testday int
		var practice int
		var qual int
		var warnup int
		var race int
		err := rows.Scan(&testday, &practice, &qual, &warnup, &race)
		if err != nil {
			return n, err
		}
		n.setSessionTypeEnabledFlag(TestDay, testday == 1)
		n.setSessionTypeEnabledFlag(Practice, practice == 1)
		n.setSessionTypeEnabledFlag(Qual, qual == 1)
		n.setSessionTypeEnabledFlag(Warmup, warnup == 1)
		n.setSessionTypeEnabledFlag(Race, race == 1)
		return n, nil
	}
	err := rows.Err()
	if err != nil {
		return n, err
	}
	return n, err
}

func buildSelectSessionStartedCommand(sessionType string) (string, func(rows *sql.Rows) ([]TelegramUser, error)) {
	fields := "userid, name, chatid"
	return fmt.Sprintf(`SELECT %s FROM notifications WHERE %s = 1`, fields, sessionType), processSelectSessionStartedRows
}

func processSelectSessionStartedRows(rows *sql.Rows) ([]TelegramUser, error) {
	defer rows.Close()

	users := make([]TelegramUser, 0)
	for rows.Next() {
		var id string
		var name string
		var chatid string
		err := rows.Scan(&id, &name, &chatid)
		if err != nil {
			return users, err
		}
		users = append(users, TelegramUser{
			ID:     id,
			Name:   name,
			ChatID: chatid,
		})
	}
	err := rows.Err()
	if err != nil {
		return users, err
	}
	return users, err
}

func buildUpdateUserCommand(userID, chatID string, n Notifications) string {
	testday := n.TestDayEnabledInt()
	practice := n.PracticeEnabledInt()
	qual := n.QualEnabledInt()
	warnup := n.WarmupEnabledInt()
	race := n.RaceEnabledInt()

	fields := "userid, name, chatid, testday, practice, qual, warnup, race"
	// columns := fmt.Sprintf(`testday = %d, practice = %d, qual = %d, warnup = %d, race = %d`, testday, practice, qual, warnup, race)
	values := fmt.Sprintf(`'%s', '%s', '%s', %d, %d, %d, %d, %d`, userID, userID, chatID, testday, practice, qual, warnup, race)
	// return fmt.Sprintf(`UPDATE notifications SET %s WHERE id = '%s'`, columns, userID)
	return fmt.Sprintf(`INSERT OR REPLACE INTO notifications (%s) VALUES (%s)`, fields, values)
}
