package session

import (
	"fmt"
	"reflect"
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

// First gets the 1st row
func (s *Session) First(value interface{}) error {
	table := s.RefTable(value)

	fieldSQL := strings.Join(table.FieldNames, ", ")
	selectSQL := fmt.Sprintf("SELECT %v FROM %s LIMIT 1", fieldSQL, table.TableName)
	row := s.Raw(selectSQL).QueryRow()

	dest := reflect.ValueOf(value).Elem()
	var values []interface{}
	for _, name := range table.FieldNames {
		values = append(values, dest.FieldByName(name).Addr().Interface())
	}

	return row.Scan(values...)
}

// Find gets all eligible records
func (s *Session) Find(values interface{}) error {
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	table := s.RefTable(reflect.New(destType).Elem().Interface())

	fieldSQL := strings.Join(table.FieldNames, ", ")
	selectSQL := fmt.Sprintf("SELECT %v FROM %s", fieldSQL, table.TableName)
	rows, err := s.Raw(selectSQL).QueryRows()
	if err != nil {
		return err
	}

	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var values []interface{}
		for _, name := range table.FieldNames {
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		if err := rows.Scan(values...); err != nil {
			return err
		}
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	if err := rows.Close(); err != nil {
		return err
	}
	return nil
}
