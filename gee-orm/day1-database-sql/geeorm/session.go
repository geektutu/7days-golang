package geeorm

import (
	"database/sql"
	"fmt"
	"strings"

	"geeorm/schema"
)

type Session struct {
	engine   *Engine
	refTable *schema.Schema

	Value   interface{}
	SQL     strings.Builder
	SQLVars []interface{}
}

func (s *Session) Exec() (result sql.Result, err error) {
	if result, err = s.engine.db.Exec(s.SQL.String(), s.SQLVars...); err != nil {
		ErrorLog.Println(err)
	}
	return
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
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

func (s *Session) CreateTable() *Session {
	var columns []string
	for _, field := range s.refTable.Fields {
		columns = append(columns, fmt.Sprintf("%s %s", field.Name, field.Tag))
	}
	desc := strings.Join(columns, ",")
	s.SQL.WriteString(fmt.Sprintf("CREATE TABLE %s (%s);", s.refTable.Table, desc))
	return s
}
