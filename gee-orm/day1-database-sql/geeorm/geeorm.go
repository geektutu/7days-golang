package geeorm

import (
	"database/sql"

	"geeorm/log"
	"geeorm/session"
)

// Engine is the main struct of geeorm, manages all db sessions and transactions.
type Engine struct {
	db *sql.DB
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
	e = &Engine{db: db}
	log.Info.Println("Connect database success")
	return
}

// Close database connection
func (engine *Engine) Close() (err error) {
	if err = engine.db.Close(); err == nil {
		log.Info.Println("Close database success")
	}
	return
}

// NewSession creates a new session for next operations
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db)
}
