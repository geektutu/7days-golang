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

func TestSession_First(t *testing.T) {
	_ = NewSession().DropTable(&User{})
	_ = NewSession().CreateTable(&User{})
	_, _ = NewSession().Create(user1)
	u := &User{}
	err := NewSession().First(u)
	if err != nil || u.Age != user1.Age || u.Name != user1.Name {
		t.Fatal("failed to query first")
	}
}

func TestSession_Find(t *testing.T) {
	_ = NewSession().DropTable(&User{})
	_ = NewSession().CreateTable(&User{})
	_, _ = NewSession().Create(user1, user2)
	users := []User{}
	if err := NewSession().Find(&users); err != nil || len(users) != 2 {
		t.Fatal("failed to query all")
	}
}
