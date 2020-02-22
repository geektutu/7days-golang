package geeorm

import (
	"fmt"
	"strings"

	"geeorm/schema"
)

func (s *Session) RefTable(value interface{}) *schema.Schema {
	if value == nil {
		panic("value is nil")
	}
	if s.refTable == nil {
		s.refTable = schema.Parse(value, s.engine.dialect)
	}
	return s.refTable
}

func (s *Session) CreateTable(value interface{}) error {
	table := s.RefTable(value)
	var columns []string
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s", field.Name, field.Tag))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.TableName, desc)).Exec()
	return err
}

func (s *Session) DropTable(value interface{}) error {
	table := s.RefTable(value)
	_, err := s.Raw(fmt.Sprintf("DROP TABLE %s", table.TableName)).Exec()
	return err
}

func (s *Session) HasTable(value interface{}) bool {
	tableName, ok := value.(string)
	if !ok {
		tableName = s.RefTable(value).TableName
	}

	sql, values := s.engine.dialect.TableExistSQL(tableName)
	row := s.Raw(sql, values...).QueryRow()



	var tmp string
	if err := row.Scan(&tmp); err != nil {
		ErrorLog.Println(err)
	}
	return tmp == tableName
}
