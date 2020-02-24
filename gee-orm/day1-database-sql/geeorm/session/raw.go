package session

import (
	"database/sql"
	"strings"

	"geeorm/log"
)

// Session keep a pointer to sql.DB and provides all execution of all
// kind of database operations.
type Session struct {
	db *sql.DB

	SQL     strings.Builder
	SQLVars []interface{}
}

// New creates a instance of Session
func New(db *sql.DB) *Session {
	return &Session{db: db}
}

// Exec raw SQL with SQLVars
func (s *Session) Exec() (result sql.Result, err error) {
	log.Info(s.SQL.String(), s.SQLVars)
	if result, err = s.db.Exec(s.SQL.String(), s.SQLVars...); err != nil {
		log.Error(err)
	}
	return
}

// QueryRow gets a record from db
func (s *Session) QueryRow() *sql.Row {
	log.Info(s.SQL.String(), s.SQLVars)
	return s.db.QueryRow(s.SQL.String(), s.SQLVars...)
}

// QueryRows gets a list of records from db
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	log.Info(s.SQL.String(), s.SQLVars)
	if rows, err = s.db.Query(s.SQL.String(), s.SQLVars...); err != nil {
		log.Error(err)
	}
	return
}

// Raw appends SQL and SQLVars
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.SQL.WriteString(sql)
	s.SQLVars = append(s.SQLVars, values...)
	return s
}
