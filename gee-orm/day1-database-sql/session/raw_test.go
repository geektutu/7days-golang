package session

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var TestDB *sql.DB

func TestMain(m *testing.M) {
	TestDB, _ = sql.Open("sqlite3", "gee.db")
	code := m.Run()
	_ = TestDB.Close()
	os.Exit(code)
}

func NewSession() *Session {
	return &Session{db: TestDB}
}

func TestSession_Exec(t *testing.T) {
	_, _ = NewSession().Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = NewSession().Raw("CREATE TABLE User(name text);").Exec()
	result, _ := NewSession().Raw("INSERT INTO User(`name`) values (?), (?)", "Tom", "Sam").Exec()
	if count, err := result.RowsAffected(); err != nil || count != 2 {
		t.Fatal("expect 2, but got", count)
	}
}

func TestSession_QueryRows(t *testing.T) {
	_, _ = NewSession().Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = NewSession().Raw("CREATE TABLE User(name text);").Exec()
	row := NewSession().Raw("SELECT count(*) FROM User").QueryRow()
	var count int
	if err := row.Scan(&count); err != nil || count != 0 {
		t.Fatal("failed to query db", err)
	}
}
