package schema

import (
	"geeorm/dialect"
	"testing"
)

type User struct {
	Name string
	Age  int
}

func TestParse(t *testing.T) {
	dial, _ := dialect.GetDialect("sqlite3")
	schema := Parse(&User{"Tom", 18}, dial)

	if schema.TableName != "User" || len(schema.Fields) != 2 {
		t.Fatal("failed to parse User struct")
	}
}
