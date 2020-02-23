package session

import "testing"

var (
	user1 = &User{"Tom", 18}
	user2 = &User{"Sam", 25}
)

func TestSession_Create(t *testing.T) {
	_ = NewSession().DropTable(&User{})
	_ = NewSession().CreateTable(&User{})
	if affected, err := NewSession().Create(user1, user2); err != nil || affected != 2 {
		t.Fatal("failed to create record")
	}
}
