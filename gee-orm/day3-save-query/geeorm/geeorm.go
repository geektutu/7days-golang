package geeorm

import (
	"database/sql"
	"fmt"

	"geeorm/dialect"
	"geeorm/log"
	"geeorm/session"
)

// Engine is the main struct of geeorm, manages all db sessions and transactions.
type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

// NewEngine create a instance of Engine
// connect database and ping it to test whether it's alive
func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error.Println(err)
		return
	}
	// Send a ping to make sure the database connection is alive.
	if err = db.Ping(); err != nil {
		log.Error.Println(err)
		return
	}
	// make sure the specific dialect exists
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		err = fmt.Errorf("dialect %s Not Found", driver)
		log.Error.Println(err)
		return
	}
	e = &Engine{db: db, dialect: dial}
	log.Info.Println("Connect database success")
	return
}

// Close database connection
func (e *Engine) Close() (err error) {
	if err = e.db.Close(); err == nil {
		log.Info.Println("Close database success")
	}
	return
}

// NewSession creates a new session for next operations
func (e *Engine) NewSession() *session.Session {
	return session.New(e.db, e.dialect)
}
