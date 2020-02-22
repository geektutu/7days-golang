package geeorm

import (
	"testing"
)

func TestSession_Exec(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	_, _ = engine.NewSession().Raw("DROP TABLE USER;").Exec()
	_, _ = engine.NewSession().Raw("CREATE TABLE USER(name text);").Exec()
	result, _ := engine.NewSession().
		Raw("INSERT INTO USER(`name`) values (?), (?)", "Tom", "Sam").Exec()
	if count, err := result.RowsAffected(); err != nil || count != 2 {
		t.Fatal("expect 2, but got", count)
	}
}

func TestSession_QueryRow(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	_, _ = engine.NewSession().Raw("DROP TABLE USER;").Exec()
	_, _ = engine.NewSession().Raw("CREATE TABLE USER(name text);").Exec()
	row := engine.NewSession().Raw("SELECT count(*) FROM USER").QueryRow()

	var count int
	if err := row.Scan(&count); err != nil || count != 0 {
		t.Fatal("failed to query db", err)
	}
}
