package session

import (
	"fmt"
	"strings"

	"geeorm/schema"
)

// RefTable returns a Schema instance that contains all parsed fields
func (s *Session) RefTable(value interface{}) *schema.Schema {
	if value == nil {
		panic("value is nil")
	}
	if s.refTable == nil {
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s.refTable
}

// CreateTable create a table in database with a model
func (s *Session) CreateTable(value interface{}) error {
	table := s.RefTable(value)
	var columns []string
	for _, field := range table.Fields {
		tag := field.Tag
		if field.Name == table.PrimaryField.Name {
			tag += " PRIMARY KEY"
		}
		columns = append(columns, fmt.Sprintf("%s %s", field.Name, tag))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.TableName, desc)).Exec()
	return err
}

// DropTable drops a table with the name of model
func (s *Session) DropTable(value interface{}) error {
	table := s.RefTable(value)
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", table.TableName)).Exec()
	return err
}

// HasTable returns true of the table exists
func (s *Session) HasTable(value interface{}) bool {
	tableName, ok := value.(string)
	if !ok {
		tableName = s.RefTable(value).TableName
	}

	sql, values := s.dialect.TableExistSQL(tableName)
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == tableName
}
