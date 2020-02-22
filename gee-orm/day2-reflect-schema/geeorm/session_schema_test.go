package geeorm

import "testing"

type User struct {
	Name string
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	_ = engine.NewSession().DropTable(&User{})
	_ = engine.NewSession().CreateTable(&User{})
	if ! engine.NewSession().HasTable("User") {
		t.Fatal("failed to create table User")
	}
}
