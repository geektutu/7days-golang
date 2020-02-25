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
	orders := []Type{SELECT, LIMIT}
	sql, vars := clause.Build(orders)
	t.Log(sql, vars)
	if sql != "SELECT * FROM User LIMIT ?" {
		t.Fatal("failed to build SQL")
	}
	if !reflect.DeepEqual(vars, []interface{}{3}) {
		t.Fatal("failed to build SQLVars")
	}
}
