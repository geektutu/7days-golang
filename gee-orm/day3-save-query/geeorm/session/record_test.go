package session

import "testing"

var (
	user1 = &User{"Tom", 18}
	user2 = &User{"Sam", 25}
	user3 = &User{"Jack", 25}
)

func testRecordInit(t *testing.T) {
	t.Helper()
	err1 := NewSession().DropTable(&User{})
	err2 := NewSession().CreateTable(&User{})
	_, err3 := NewSession().Create(user1, user2)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatal("failed init test records")
	}

}

func TestSession_Create(t *testing.T) {
	testRecordInit(t)
	affected, err := NewSession().Create(user3)
	if err != nil || affected != 1 {
		t.Fatal("failed to create record")
	}
}

func TestSession_Find(t *testing.T) {
	testRecordInit(t)
	users := []User{}
	if err := NewSession().Find(&users); err != nil || len(users) != 2 {
		t.Fatal("failed to query all")
	}
}
