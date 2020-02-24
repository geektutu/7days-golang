package session

import (
	"geeorm/clause"
	"reflect"
	"strings"
)

// Create one or more records in database
func (s *Session) Create(values ...interface{}) (int64, error) {
	var flag bool
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		table := s.RefTable(value)
		if !flag {
			flag = true
			fieldSQL := strings.Join(table.FieldNames, ", ")
			s.clause.Set(clause.INSERT, table.TableName, fieldSQL)
		}
		recordValues = append(recordValues, table.Values(value))
	}

	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build([]clause.Type{clause.INSERT, clause.VALUES})
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// First gets the 1st row
func (s *Session) First(value interface{}) error {
	table := s.RefTable(value)

	fieldSQL := strings.Join(table.FieldNames, ", ")

	s.clause.Set(clause.SELECT, table.TableName, fieldSQL)
	s.clause.Set(clause.LIMIT, 1)

	sql, vars := s.clause.Build([]clause.Type{clause.SELECT, clause.LIMIT})
	row := s.Raw(sql, vars...).QueryRow()

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
	s.clause.Set(clause.SELECT, table.TableName, fieldSQL)
	sql, vars := s.clause.Build([]clause.Type{clause.SELECT})
	rows, err := s.Raw(sql, vars...).QueryRows()
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
