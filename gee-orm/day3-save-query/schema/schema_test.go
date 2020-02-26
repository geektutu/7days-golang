package schema

import (
	"geeorm/dialect"
	"testing"
)

type User struct {
	Name string `geeorm:"primary_key"`
	Age  int
}

var TestDial, _ = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse(&User{}, TestDial)
	if schema.TableName != "User" || len(schema.Fields) != 2 {
		t.Fatal("failed to parse User struct")
	}
	if schema.PrimaryField.Name != "Name" {
		t.Fatal("failed to parse primary key")
	}
	t.Log(schema)
}

func TestSchema_Values(t *testing.T) {
	schema := Parse(&User{}, TestDial)
	values := schema.Values(&User{"Tom", 18})

	name := values[0].(string)
	age := values[1].(int)

	if name != "Tom" || age != 18 {
		t.Fatal("failed to get values")
	}
}
