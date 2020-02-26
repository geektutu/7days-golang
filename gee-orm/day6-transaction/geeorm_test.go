package geeorm

import (
	"errors"
	"geeorm/session"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("sqlite3", "gee.db")
	if err != nil {
		t.Fatal("failed to connect", err)
	}
	return engine
}

func CloseDB(engine *Engine) {
	_ = engine.Close()
}

func TestNewEngine(t *testing.T) {
	engine := OpenDB(t)
	_ = engine.Close()
}

type User struct {
	Name string `geeorm:"primary_key"`
	Age  int
}

func transactionRollback(t *testing.T) {
	engine := OpenDB(t)
	defer CloseDB(engine)
	s := engine.NewSession()
	_ = s.DropTable(&User{})
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.CreateTable(&User{})
		_, err = s.Create(&User{"Tom", 18})
		return nil, errors.New("Error")
	})
	if err == nil || s.HasTable("User") {
		t.Fatal("failed to rollback")
	}
}

func transactionCommit(t *testing.T) {
	engine := OpenDB(t)
	defer CloseDB(engine)
	s := engine.NewSession()
	_ = s.DropTable(&User{})
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		err = s.CreateTable(&User{})
		_, err = s.Create(&User{"Tom", 18})
		return
	})
	u := &User{}
	_ = s.First(u)
	if err != nil || u.Name != "Tom" {
		t.Fatal("failed to commit")
	}
}

func TestEngine_Transaction(t *testing.T) {
	t.Run("rollback", func(t *testing.T) {
		transactionRollback(t)
	})
	t.Run("commit", func(t *testing.T) {
		transactionCommit(t)
	})
}
