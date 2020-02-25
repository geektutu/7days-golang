package clause

import (
	"reflect"
	"testing"
)

func TestClause_Set(t *testing.T) {
	var clause Clause
	clause.Set(INSERT, "User", []string{"Name", "Age"})
	sql := clause.sql[INSERT]
	vars := clause.sqlVars[INSERT]
	t.Log(sql, vars)
	if sql != "INSERT INTO User (Name,Age)" || len(vars) != 0 {
		t.Fatal("failed to get clause")
	}
}

func TestClause_Build(t *testing.T) {
	var clause Clause
	clause.Set(LIMIT, 3)
	clause.Set(SELECT, "User", []string{"*"})
	clause.Set(WHERE, "Name = ?", 18)
	clause.Set(ORDERBY, "Age ASC")
	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	t.Log(sql, vars)
	if sql != "SELECT * FROM User WHERE Name = ? ORDER BY Age ASC LIMIT ?" {
		t.Fatal("failed to build SQL")
	}
	if !reflect.DeepEqual(vars, []interface{}{18, 3}) {
		t.Fatal("failed to build SQLVars")
	}
}
