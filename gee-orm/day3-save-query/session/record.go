package session

import (
	"geeorm/clause"
	"reflect"
)

// Insert one or more records in database
func (s *Session) Insert(values ...interface{}) (int64, error) {
	if len(values) < 1 {
		panic("There is no value to insert")
	}
	recordValues := make([]interface{}, 0)

	baseValue := values[0]
	table := s.Model(baseValue).RefTable()
	baseType := reflect.TypeOf(baseValue)
	s.clause.Set(clause.INSERT, table.Name, table.FieldNames)

	for _, value := range values {
		if reflect.TypeOf(value) != baseType {
			panic("All insert values must be the same type")
		}
		recordValues = append(recordValues, table.RecordValues(value))
	}

	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Find gets all eligible records
func (s *Session) Find(values interface{}) error {
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
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
	return rows.Close()
}
