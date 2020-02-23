package session

import (
	"testing"
)

type User struct {
	Name string `geeorm:"primary_key"`
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	_ = NewSession().DropTable(&User{})
	_ = NewSession().CreateTable(&User{})
	if !NewSession().HasTable("User") {
		t.Fatal("failed to create table User")
	}
}
