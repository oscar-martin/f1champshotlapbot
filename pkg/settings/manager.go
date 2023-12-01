package settings

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DbName = "./livetiming-bot.db"

	TestDay  = "TestDay"
	Practice = "Practice"
	Qual     = "Qual"
	Warmup   = "Warmup"
	Race     = "Race"
)

type TelegramUser struct {
	ID     string
	Name   string
	ChatID string
}

type Notifications map[string]bool

func AllEnabled() Notifications {
	return Notifications{
		TestDay:  true,
		Practice: true,
		Qual:     true,
		Warmup:   true,
		Race:     true,
	}
}

func AllDisabled() Notifications {
	return Notifications{
		TestDay:  false,
		Practice: false,
		Qual:     false,
		Warmup:   false,
		Race:     false,
	}
}

func (n Notifications) TestDaySymbol() string {
	return symbolStatus(n[TestDay])
}

func (n Notifications) PracticeSymbol() string {
	return symbolStatus(n[Practice])
}

func (n Notifications) QualSymbol() string {
	return symbolStatus(n[Qual])
}

func (n Notifications) WarmupSymbol() string {
	return symbolStatus(n[Warmup])
}

func (n Notifications) RaceSymbol() string {
	return symbolStatus(n[Race])
}

func (n Notifications) TestDayEnabledInt() int {
	if n[TestDay] {
		return 1
	}
	return 0
}

func (n Notifications) PracticeEnabledInt() int {
	if n[Practice] {
		return 1
	}
	return 0
}

func (n Notifications) QualEnabledInt() int {
	if n[Qual] {
		return 1
	}
	return 0
}

func (n Notifications) WarmupEnabledInt() int {
	if n[Warmup] {
		return 1
	}
	return 0
}

func (n Notifications) RaceEnabledInt() int {
	if n[Race] {
		return 1
	}
	return 0
}

func (n Notifications) String() string {
	status := []string{}
	status = append(status, fmt.Sprintf("%s Notificación inicio de %q", symbolStatus(n[TestDay]), TestDay))
	status = append(status, fmt.Sprintf("%s Notificación inicio de %q", symbolStatus(n[Practice]), Practice))
	status = append(status, fmt.Sprintf("%s Notificación inicio de %q", symbolStatus(n[Qual]), Qual))
	status = append(status, fmt.Sprintf("%s Notificación inicio de %q", symbolStatus(n[Warmup]), Warmup))
	status = append(status, fmt.Sprintf("%s Notificación inicio de %q", symbolStatus(n[Race]), Race))
	return strings.Join(status, "\n")
}

func symbolStatus(enabled bool) string {
	if enabled {
		return "🔔"
	}
	return "🔕"
}

func (n *Notifications) setSessionTypeEnabledFlag(sessionType string, enabled bool) {
	(*n)[sessionType] = enabled
}

type Manager struct {
	db *sql.DB
	mu sync.Mutex
}

func NewManager() (*Manager, error) {
	db, err := sql.Open("sqlite3", DbName)
	if err != nil {
		log.Printf("error opening database: %s\n", err)
		return nil, err
	}

	initTableStmt := buildCreateNotificationsTable()

	_, err = db.Exec(initTableStmt)
	if err != nil {
		log.Printf("error init database: %s\n", err)
		return nil, err
	}

	return &Manager{
		db: db,
		mu: sync.Mutex{},
	}, nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.db.Close()
}

func (m *Manager) ToggleNotificationForSessionStarted(userID, chatID, sessionType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	n, err := m.listNotificationsForSessionStarted(userID)
	if err != nil {
		return err
	}

	n.setSessionTypeEnabledFlag(sessionType, !n[sessionType])
	_, err = m.db.Exec(buildUpdateUserCommand(userID, chatID, n))
	if err != nil {
		log.Printf("error updating database: %s\n", err)
		return err
	}
	return nil
}

func (m *Manager) ListNotifications(userID string) (Notifications, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.listNotificationsForSessionStarted(userID)
}

func (m *Manager) ListUsersForSessionStarted(sessionType string) ([]TelegramUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	users := []TelegramUser{}
	sql, read := buildSelectSessionStartedCommand(sessionType)
	rows, err := m.db.Query(sql)
	if err != nil {
		return users, err
	}
	return read(rows)
}

func (m *Manager) listNotificationsForSessionStarted(userID string) (Notifications, error) {
	n := AllDisabled()

	sql, read := buildSelectUserCommand(userID)
	rows, err := m.db.Query(sql)
	if err != nil {
		return n, err
	}
	return read(rows)
}
