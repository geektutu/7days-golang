package session

import (
	"fmt"
	"strings"
)

// Create one or more records in database
func (s *Session) Create(values ...interface{}) (int64, error) {
	var flag bool
	for i, value := range values {
		table := s.RefTable(value)
		filedSQL := strings.Join(table.FieldNames, ", ")
		bindVarSQL := strings.Join(table.BindVars, ", ")
		if !flag {
			s.Raw(fmt.Sprintf("INSERT INTO %s (%v) VALUES ", table.TableName, filedSQL))
			flag = true
		}
		s.Raw(fmt.Sprintf("(%v)", bindVarSQL), table.Values(value)...)
		if i == len(values)-1 {
			s.Raw(";")
		} else {
			s.Raw(",")
		}
	}

	result, err := s.Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (s *Session) First(value interface{}) error {
	table := s.RefTable(value)
	fieldSQL := strings.Join(table.FieldNames, ", ")
	sql := fmt.Sprintf("SELECT (%v) FROM %s LIMIT 1", fieldSQL, table.TableName)

	row := s.Raw(sql).QueryRow()

	return row.Scan(value)
}
