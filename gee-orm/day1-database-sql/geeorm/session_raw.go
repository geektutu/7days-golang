package geeorm

import (
	"database/sql"
	"strings"
)

type Session struct {
	engine *Engine

	SQL     strings.Builder
	SQLVars []interface{}
}

func (s *Session) Exec() (result sql.Result, err error) {
	InfoLog.Println(s.SQL, s.SQLVars)
	if result, err = s.engine.db.Exec(s.SQL.String(), s.SQLVars...); err != nil {
		ErrorLog.Println(err)
	}
	return
}

func (s *Session) QueryRow() *sql.Row {
	InfoLog.Println(s.SQL, s.SQLVars)
	return s.engine.db.QueryRow(s.SQL.String(), s.SQLVars...)
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	InfoLog.Println(s.SQL, s.SQLVars)
	if rows, err = s.engine.db.Query(s.SQL.String(), s.SQLVars...); err != nil {
		ErrorLog.Println(err)
	}
	return
}

func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.SQL.WriteString(sql)
	s.SQLVars = values
	return s
}
