package session

import (
	"testing"
)

type User struct {
	Name string
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	s := NewSession()
	_ = s.DropTable(&User{})
	_ = s.CreateTable(&User{})
	if !s.HasTable("User") {
		t.Fatal("failed to create table User")
	}
}
