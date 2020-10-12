package schema

import (
	"geeorm/dialect"
	"testing"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

var TestDial, _ = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse(&User{}, TestDial)
	if schema.Name != "User" || len(schema.Fields) != 2 {
		t.Fatal("failed to parse User struct")
	}
	if schema.GetField("Name").Tag != "PRIMARY KEY" {
		t.Fatal("failed to parse primary key")
	}
}

func TestSchema_RecordValues(t *testing.T) {
	schema := Parse(&User{}, TestDial)
	values := schema.RecordValues(&User{"Tom", 18})

	name := values[0].(string)
	age := values[1].(int)

	if name != "Tom" || age != 18 {
		t.Fatal("failed to get values")
	}
}

type UserTest struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func (u *UserTest) TableName() string {
	return "ns_user_test"
}

func TestSchema_TableName(t *testing.T) {
	schema := Parse(&UserTest{}, TestDial)
	if schema.Name != "ns_user_test" || len(schema.Fields) != 2 {
		t.Fatal("failed to parse User struct")
	}
}
