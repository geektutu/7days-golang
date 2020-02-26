package session

import (
	"geeorm/clause"
	"reflect"
)

// Create one or more records in database
func (s *Session) Create(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		table := s.RefTable(value)
		s.clause.Set(clause.INSERT, table.TableName, table.FieldNames)
		recordValues = append(recordValues, table.Values(value))
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
	table := s.RefTable(reflect.New(destType).Elem().Interface())

	s.clause.Set(clause.SELECT, table.TableName, table.FieldNames)
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

// First gets the 1st row
func (s *Session) First(value interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	err := s.Limit(1).Find(destSlice.Addr().Interface())
	dest.Set(destSlice.Index(0))
	return err
}

// Limit adds limit condition to clause
func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

// Where adds limit condition to clause
func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

// OrderBy adds order by condition to clause
func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

// Set adds Assignment by condition to clause
// support map[string]interface{}
// also support "Name", "Tom", "Age", 18, etc
func (s *Session) Set(values ...interface{}) *Session {
	m, ok := values[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(values); i += 2 {
			m[values[i].(string)] = values[i+1]
		}
	}
	s.clause.Set(clause.SET, m)
	return s
}

// Update records with where clause
func (s *Session) Update(value interface{}) (int64, error) {
	s.clause.Set(clause.UPDATE, s.guessTableName(value))
	sql, vars := s.clause.Build(clause.UPDATE, clause.SET, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete records with where clause
func (s *Session) Delete(value interface{}) (int64, error) {
	s.clause.Set(clause.DELETE, s.guessTableName(value))
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Count records with where clause
func (s *Session) Count(value interface{}) (int64, error) {
	s.clause.Set(clause.COUNT, s.guessTableName(value))
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var tmp int64

	if err := row.Scan(&tmp); err != nil {
		return 0, err
	}
	return tmp, nil
}
