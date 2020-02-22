package geeorm

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"geeorm/dialect"
	"geeorm/schema"
)

var (
	ErrorLog = log.New(os.Stdout, "[error] ", log.LstdFlags)
	InfoLog  = log.New(os.Stdout, "[info ] ", log.LstdFlags)
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		ErrorLog.Println(err)
		return
	}
	// Send a ping to make sure the database connection is alive.
	if err = db.Ping(); err != nil {
		ErrorLog.Println(err)
		return
	}
	// make sure the specific dialect exists
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		err = fmt.Errorf("dialect %s Not Found", driver)
		ErrorLog.Println(err)
		return
	}
	e = &Engine{db: db, dialect: dial}
	InfoLog.Println("Connect database success")
	return
}

func (e *Engine) Close() (err error) {
	if err = e.db.Close(); err == nil {
		InfoLog.Println("Close database success")
	}
	return
}

func (e *Engine) CreateTable(value interface{}) error {
	_, err := e.NewSession(value).CreateTable().Exec()
	return err
}

func (e *Engine) NewSession(value interface{}) *Session {
	var refTable *schema.Schema
	if value != nil {
		refTable = schema.Parse(value, e.dialect)
	}
	return &Session{
		refTable: refTable,
		engine:   e,
	}
}
